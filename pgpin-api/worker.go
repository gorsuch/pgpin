package main

import (
	"bitbucket.org/kardianos/table"
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pq"
	"log"
	"time"
)

func workerPoll() (*pin, error) {
	pin, err := dataPinForQuery()
	if err != nil {
		return nil, err
	} else if pin != nil {
		log.Printf("worker.poll.found pin_id=%s", pin.Id)
		return pin, nil
	}
	return nil, nil
}

func workerQuery(p *pin) error {
	log.Printf("worker.query.start pin_id=%s", p.Id)
	dbUrl, err := dataPinDbUrl(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.reserve pin_id=%s", p.Id)
	startedAt := time.Now()
	p.QueryStartedAt = &startedAt
	p, err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.open pin_id=%s", p.Id)
	resourceDb, err := sql.Open("postgres", dataMustParseDatabaseUrl(dbUrl))
	if err != nil {
		return err
	}
	log.Printf("worker.query.exec pin_id=%s", p.Id)
	buffer, err := table.Get(resourceDb, p.Query)
	finishedAt := time.Now()
	p.QueryFinishedAt = &finishedAt
	if err != nil {
		if pgerr, ok := err.(pq.PGError); ok {
			log.Printf("worker.query.usererror pin_id=%s", p.Id)
			msg := pgerr.Get('M')
			p.ResultsError = &msg
			p, err = dataPinUpdate(p)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	log.Printf("worker.query.read pin_id=%s", p.Id)
	resultsFieldsJsonB, _ := json.Marshal(buffer.ColumnName)
	resultsFieldsJson := string(resultsFieldsJsonB)
	resultsRows := make([][]interface{}, len(buffer.Rows))
	for i, row := range buffer.Rows {
		resultsRows[i] = make([]interface{}, len(row.Data))
		for j, rowDatum := range row.Data {
			switch rowValue := rowDatum.(type) {
			case []byte:
				resultsRows[i][j] = string(rowValue)
			default:
				resultsRows[i][j] = rowValue
			}
		}
	}
	resultsRowsJsonB, _ := json.Marshal(resultsRows)
	resultsRowsJson := string(resultsRowsJsonB)
	log.Printf("worker.query.commit pin_id=%s", p.Id)
	p.ResultsFieldsJson = &resultsFieldsJson
	p.ResultsRowsJson = &resultsRowsJson
	p, err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.finish pin_id=%s", p.Id)
	return nil
}

func workerTick() {
	pin, err := workerPoll()
	if err != nil {
		panic(err)
	}
	if pin != nil {
		err = workerQuery(pin)
		if err != nil {
			panic(err)
		}
	} else {
		time.Sleep(time.Millisecond * 250)
	}
}

func workerStart() {
	log.Print("worker.start")
	for {
		workerTick()
	}
}
