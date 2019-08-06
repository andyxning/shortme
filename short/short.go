package short

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"

	"github.com/andyxning/shortme/base"
	"github.com/andyxning/shortme/conf"
	"github.com/andyxning/shortme/sequence"
	_ "github.com/andyxning/shortme/sequence/db"
	"github.com/go-sql-driver/mysql"
)

type shorter struct {
	readDB   *sql.DB
	writeDB  *sql.DB
	sequence sequence.Sequence
}

// connect will panic when it can not connect to DB server.
func (shorter *shorter) mustConnect() {
	shorter.reconnectReadDB()
	shorter.reconnectWriteDB()
}

func (shorter *shorter) reconnectReadDB() {
	if shorter.readDB != nil {
		shorter.readDB.Close()
		shorter.readDB = nil
	}

	db, err := sql.Open("mysql", conf.Conf.ShortDB.ReadDSN)
	if err != nil {
		log.Panicf("short read db open error. %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("short read db ping error. %v", err)
	}

	db.SetMaxIdleConns(conf.Conf.ShortDB.MaxIdleConns)
	db.SetMaxOpenConns(conf.Conf.ShortDB.MaxOpenConns)

	shorter.readDB = db
}

func (shorter *shorter) reconnectWriteDB() {
	if shorter.writeDB != nil {
		shorter.writeDB.Close()
		shorter.writeDB = nil
	}

	db, err := sql.Open("mysql", conf.Conf.ShortDB.WriteDSN)
	if err != nil {
		log.Panicf("short write db open error. %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("short write db ping error. %v", err)
	}

	db.SetMaxIdleConns(conf.Conf.ShortDB.MaxIdleConns)
	db.SetMaxOpenConns(conf.Conf.ShortDB.MaxOpenConns)

	shorter.writeDB = db
}

// initSequence will panic when it can not open the sequence successfully.
func (shorter *shorter) mustInitSequence() {
	shorter.reconnectSequence()
}

func (shorter *shorter) reconnectSequence() {
	if shorter.sequence != nil {
		shorter.sequence.Close()
		shorter.sequence = nil
	}

	sequence, err := sequence.GetSequence("db")
	if err != nil {
		log.Panicf("get sequence instance error. %v", err)
	}

	err = sequence.Open()
	if err != nil {
		log.Panicf("open sequence instance error. %v", err)
	}

	shorter.sequence = sequence
}

func (shorter *shorter) close() {
	if shorter.readDB != nil {
		shorter.readDB.Close()
		shorter.readDB = nil
	}

	if shorter.writeDB != nil {
		shorter.writeDB.Close()
		shorter.writeDB = nil
	}
}

func (shorter *shorter) connectionError(err *error) bool {
	return *err == driver.ErrBadConn || *err == mysql.ErrInvalidConn
}

func (shorter *shorter) Expand(shortURL string) (longURL string, err error) {
	selectSQL := fmt.Sprintf(`SELECT long_url FROM short WHERE short_url=?`)

	var rows *sql.Rows
	rows, err = shorter.readDB.Query(selectSQL, shortURL)

	if err != nil {
		if shorter.connectionError(&err) {
			shorter.reconnectReadDB()

			rows, err = shorter.readDB.Query(selectSQL, shortURL)
			if err != nil {
				log.Printf("short read db query error. %v", err)
				return "", errors.New("short read db query error")
			}
		} else {
			log.Printf("short read db query error. %v", err)
			return "", errors.New("short read db query error")
		}
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&longURL)
		if err != nil {
			log.Printf("short read db query rows scan error. %v", err)
			return "", errors.New("short read db query rows scan error")
		}
	}

	err = rows.Err()
	if err != nil {
		log.Printf("short read db query rows iterate error. %v", err)
		return "", errors.New("short read db query rows iterate error")
	}

	return longURL, nil
}

func (shorter *shorter) Short(longURL string) (shortURL string, err error) {
	for {
		var seq uint64
		seq, err = shorter.sequence.NextSequence()
		if err != nil {
			if shorter.connectionError(&err) {
				shorter.reconnectSequence()

				seq, err = shorter.sequence.NextSequence()
				if err != nil {
					log.Printf("get next sequence error. %v", err)
					return "", errors.New("get next sequence error")
				}
			} else {
				log.Printf("get next sequence error. %v", err)
				return "", errors.New("get next sequence error")
			}
		}

		shortURL = base.Int2String(seq)
		if _, exists := conf.Conf.Common.BlackShortURLsMap[shortURL]; exists {
			continue
		} else {
			break
		}
	}

	insertSQL := fmt.Sprintf(`INSERT INTO short(long_url, short_url) VALUES(?, ?)`)

	var stmt *sql.Stmt
	stmt, err = shorter.writeDB.Prepare(insertSQL)
	if err != nil {
		if shorter.connectionError(&err) {
			shorter.reconnectWriteDB()

			stmt, err = shorter.writeDB.Prepare(insertSQL)
			if err != nil {
				log.Printf("short write db prepares error. %v", err)
				return "", errors.New("short write db prepares error")
			}
		} else {
			log.Printf("short write db prepares error. %v", err)
			return "", errors.New("short write db prepares error")
		}
	}
	defer stmt.Close()

	_, err = stmt.Exec(longURL, shortURL)
	if err != nil {
		if shorter.connectionError(&err) {
			shorter.reconnectWriteDB()

			_, err = stmt.Exec(longURL, shortURL)
			if err != nil {
				log.Printf("short write db insert error. %v", err)
				return "", errors.New("short write db insert error")
			}
		} else {
			log.Printf("short write db insert error. %v", err)
			return "", errors.New("short write db insert error")
		}
	}

	return shortURL, nil
}

var Shorter shorter

func Start() {
	Shorter.mustConnect()
	Shorter.mustInitSequence()
	log.Println("shorter starts")
}
