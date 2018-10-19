package server
import(
	"reflect"
	"database/sql"
	"fmt"
	"strings"
)

type tag struct {
	Id int
	Name string `json:"name"`
}
type Entry struct {
	Id	 int64
	Title	string `json:"title"`
	Url string `json:"url"`
	BaseTime int64 `json:"baseTime"`
	BeginTime int64 `json:"beginTime"`
	EndTime int64 `json:"endTime"`
	Site	int64 `json:"site"`
	Tags []*tag
}

func (self *Entry) LoadDB(id int64,db *sql.DB) (err error) {

	row := db.QueryRow("SELECT id,title,url,baseTime,beginTime,endTime,site FROM entry WHERE id = ?",id)
	err = row.Scan(
		&self.Id,
		&self.Title,
		&self.Url,
		&self.BaseTime,
		&self.BeginTime,
		&self.EndTime,
		&self.Site)
	if err != nil {
		return err
	}
	return nil

}

func ReadEntry(hand func(*Entry) error,where string,val ...interface{}) error {
	sql_ := "SELECT id,title,url,baseTime,beginTime,endTime,site FROM entry " + where + ";"
	var en Entry
	return HandDBForBack(Conf.DbPath,func(db *sql.DB)error{
		row,err := db.Query(sql_,val)
		if err != nil {
			return err
		}
		for row.Next() {
			err = row.Scan(
				&en.Id,
				&en.Title,
				&en.Url,
				&en.BaseTime,
				&en.BeginTime,
				&en.EndTime,
				&en.Site)
			if err != nil {
				return err
			}
			err = hand(&en)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func StructSaveForDB(db *sql.DB,TableName string,st interface{}) sql.Result {
	re := reflect.TypeOf(st).Elem()
	va := reflect.ValueOf(st).Elem()
	var fi []string
	var x []string
	var val []interface{}
	for i:=0;i<re.NumField();i++ {
		str :=re.Field(i).Tag.Get("json")
		if str != "" {
			v:= va.Field(i).Interface()
			if v  != nil{
				x = append(x,"?")
				fi = append(fi,str)
				val = append(val,v)
			}
		}
	}
	sql_ := fmt.Sprintf("INSERT INTO %s (%s) values (%s)",TableName,strings.Join(fi,","),strings.Join(x,","))
	//fmt.Println(sql_)
	res,err := db.Exec(sql_,val...)
	if err != nil {
		panic(err)
	}
	return res
	//__id,__err :=m.RowsAffected()

	//fmt.Println(_id,_err,__id,__err)
}

func StructUpdateForDB(db *sql.DB,TableName string,st interface{},keyname string ,key interface{}) sql.Result {
	re := reflect.TypeOf(st).Elem()
	va := reflect.ValueOf(st).Elem()
	var fi []string
	var val []interface{}
	for i:=0;i<re.NumField();i++ {
		str :=re.Field(i).Tag.Get("json")
		if str != "" {
			v:= va.Field(i).Interface()
			if v  != nil{
				fi = append(fi,str+" = ?")
				val = append(val,v)
			}
		}
	}
	val = append(val,key)
	sql_ := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",TableName,strings.Join(fi,","),keyname)
	if db == nil {
		fmt.Println(sql_)

		panic("db == nil")
	}
	res,err := db.Exec(sql_,val...)
	if err != nil {
		panic(err)
	}
	return res

}
