package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/builderx"
)

var (
	airbnbInfoFieldNames          = builderx.RawFieldNames(&AirbnbInfo{})
	airbnbInfoRows                = strings.Join(airbnbInfoFieldNames, ",")
	airbnbInfoRowsExpectAutoSet   = strings.Join(stringx.Remove(airbnbInfoFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	airbnbInfoRowsWithPlaceHolder = strings.Join(stringx.Remove(airbnbInfoFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"
)

type (
	AirbnbInfoModel interface {
		Insert(data AirbnbInfo) (sql.Result, error)
		FindOne(id int64) (*AirbnbInfo, error)
		Update(data AirbnbInfo) error
		Delete(id int64) error
	}

	defaultAirbnbInfoModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AirbnbInfo struct {
		AirbnbUrl        string    `db:"airbnb_url"`
		Price            float64   `db:"price"`
		CheckInDate      time.Time `db:"check_in_date"`
		CheckOutDate     time.Time `db:"check_out_date"`
		Id               int64     `db:"id"`
		HotelName        string    `db:"hotel_name"`
		Star             int64     `db:"star"`
		PriceBeforeTaxes float64   `db:"price_before_taxes"`
		Guests           int64     `db:"guests"`
		CreatedAt        time.Time `db:"created_at"`
	}
)

func NewAirbnbInfoModel(conn sqlx.SqlConn) AirbnbInfoModel {
	return &defaultAirbnbInfoModel{
		conn:  conn,
		table: "`airbnb_info`",
	}
}

func (m *defaultAirbnbInfoModel) Insert(data AirbnbInfo) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, airbnbInfoRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.AirbnbUrl, data.Price, data.CheckInDate, data.CheckOutDate, data.HotelName, data.Star, data.PriceBeforeTaxes, data.Guests, data.CreatedAt)
	return ret, err
}

func (m *defaultAirbnbInfoModel) FindOne(id int64) (*AirbnbInfo, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", airbnbInfoRows, m.table)
	var resp AirbnbInfo
	err := m.conn.QueryRow(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultAirbnbInfoModel) Update(data AirbnbInfo) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, airbnbInfoRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.AirbnbUrl, data.Price, data.CheckInDate, data.CheckOutDate, data.HotelName, data.Star, data.PriceBeforeTaxes, data.Guests, data.CreatedAt, data.Id)
	return err
}

func (m *defaultAirbnbInfoModel) Delete(id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}
