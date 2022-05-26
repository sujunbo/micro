package db

import (
	"fmt"
	"github.com/sujunbo/micro/log"
	"strings"
	"time"
)

type sqlEventReceiver struct {
	costThreshold int64
}

func NewEventReceiver(dbname string, costThreshold int64) *sqlEventReceiver {
	return &sqlEventReceiver{}
}

// Event receives a simple notification when various events occur
func (s *sqlEventReceiver) Event(eventName string) {
	log.Infof("DB Event name %v", eventName)
}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (s *sqlEventReceiver) EventKv(eventName string, kvs map[string]string) {
	log.Infof("DB EventKv name %v kv %v", eventName, kvs)
}

// EventErr receives a notification of an error if one occurs
func (s *sqlEventReceiver) EventErr(eventName string, err error) error {
	log.Errorf("DB EventErr name:%v err:%v", eventName, err)
	return err
}

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (s *sqlEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	log.Errorf("DB EventErr name:%v err:%v kvs:%v", eventName, err, kvs)
	return err
}

// Timing receives the time an event took to happen
func (s *sqlEventReceiver) Timing(eventName string, nanoseconds int64) {
	t := int64(time.Duration(nanoseconds) / time.Millisecond)
	if t > s.costThreshold {
		log.Infof("DB Timing name:%v cost:%v", eventName, time.Duration(nanoseconds).String())
	}
}

// TimingKv receives the time an event took to happen along with optional key/value data
func (s *sqlEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	t := int64(time.Duration(nanoseconds) / time.Millisecond)
	if t > s.costThreshold {
		log.Infof("DB TimingKv name:%v kv:%v cost:%v", eventName, kvs, time.Duration(nanoseconds).String())
	}
}

// SELECT * FROM {table} WHERE
// UPDATE `push_data_tab_20200401` SET `push_flag` = 1 WHERE (`user_id` IN (442547)) AND (`biz_id` = 'dc5d7e5b0efa438d97f466d66257b121')
// INSERT INTO `crm_shop_attach_tab` (`id`,`shop_id`,`calculate_buyer_time`,`buyer_num`,`extra`,`is_delete`,`ctime`,`mtime`) VALUES (0,439510,1586102400,9,'',0,1585735589,1585735589)
func table(query string) (name string) {
	qs := strings.Split(query, " ")
	for i, s := range qs {
		if s == "FROM" && i < len(qs)-1 {
			return qs[i+1]
		}
	}

	for i, s := range qs {
		if s == "UPDATE" && i < len(qs)-1 {
			return strings.Replace(qs[i+1], "`", "", -1)
		}
	}

	for i, s := range qs {
		if s == "INSERT" && i < len(qs)-2 && qs[i+1] == "INTO" {
			return strings.Replace(qs[i+2], "`", "", -1)
		}
	}

	return query
}

func dbName(dataSource string) string {
	idx := strings.Index(dataSource, "/")
	if idx == -1 {
		panic(fmt.Sprintf("datasource err:%v", dataSource))
	}

	dataSource = dataSource[idx+1:]
	idx = strings.Index(dataSource, "?")
	if idx == -1 {
		panic(fmt.Sprintf("datasource err:%v", dataSource))
	}
	return dataSource[:idx]

}
