package messages

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/lerenn/log"
	"github.com/lerenn/telerdd-server/src/api/v1/account"
	"github.com/lerenn/telerdd-server/src/data"
	"github.com/lerenn/telerdd-server/src/tools"
)

type Message struct {
	// Infos
	data   *data.Data
	db     *sql.DB
	logger *log.Log
}

func NewMessage(data *data.Data, db *sql.DB, logger *log.Log) *Message {
	var m Message
	m.data = data
	m.db = db
	m.logger = logger
	return &m
}

func (m *Message) Process(r *http.Request) string {
	switch r.Method {
	case "GET":
		return m.Get(r)
	case "POST":
		return tools.JSONError("Method not implemented")
	case "PUT":
		return m.Put(r)
	case "DELETE":
		return tools.JSONError("Method not implemented")
	default:
		return tools.JSONError("Unknown HTTP Method")
	}
}

func (m *Message) Get(r *http.Request) string {
	// Get id
	id, _, err := tools.GetIntFromRequest(r, "id")
	if err != nil {
		return tools.JSONError("Invalid number for id")
	} else if id < 0 {
		errStr := fmt.Sprintf("Invalid number for id (%d)", id)
		return tools.JSONError(errStr)
	}

	// Get complete list
	sqlReq := fmt.Sprintf("SELECT id,message,time,name,status FROM messages WHERE id = %d", id)
	rows, err := m.db.Query(sqlReq)
	if err != nil {
		return tools.JSONError(err.Error())
	}

	if rows.Next() {
		// Get infos
		var id int
		var txt, time, name, status string
		if err := rows.Scan(&id, &txt, &time, &name, &status); err != nil {
			return tools.JSONError(err.Error())
		}

		// Treat time to convenient format
		time, err := tools.SQLFormatDateTime(time)
		if err != nil {
			time = err.Error()
		}

		return messageToJSON(id, txt, time, name, status)
	}

	return tools.JSONError("No message corresponding to this Message")
}

func (m *Message) Put(r *http.Request) string {
	// Check permission
	errAuth := account.Authorized(m.db, r, 1)
	if errAuth != nil {
		return tools.JSONError(errAuth.Error())
	}

	// Get id
	id, _, err := tools.GetIntFromRequest(r, "id")
	if err != nil {
		return tools.JSONError("Invalid number for id")
	} else if id < 0 {
		id = 0
	}

	// Get status
	status := r.FormValue("status")
	if status != "accepted" && status != "refused" {
		return tools.JSONError("Invalid status")
	}

	// Prepare request
	stmt, errPrep := m.db.Prepare("UPDATE messages SET status=? WHERE id=?")
	if errPrep != nil {
		return tools.JSONError(errPrep.Error())
	}

	_, errExec := stmt.Exec(status, id)
	if errExec != nil {
		return tools.JSONError(errExec.Error())
	}

	return tools.JSONResponseOk()
}
