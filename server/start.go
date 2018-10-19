package server
import(
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/gin-gonic/gin"
	"flag"
	"io"
	"net/http"
	"compress/gzip"
	"time"
	"os"
	"log"
	"fmt"
	//"path/filepath"
)

const(
	DBType string = "sqlite3"
)

var (
	FileName   = flag.String("c", "conf.log", "config log")
	Router *gin.Engine
	EndEntry *Entry = new(Entry)
	Conf *Config
	SiteRun map[int64]*SitePage = make(map[int64]*SitePage)
	//Header   http.Header
)
type _row interface{
	Scan(dest ...interface{}) error
}
func init(){
	flag.Parse()
	Conf = NewConfig()
	_,err :=  os.Stat(Conf.DbPath)
	if err != nil {
		createDB()
	}
	if Conf.Server {
		go loadRouter()
		//go Router.Run(Conf.Port)
	}
	go runColl()
	//loadAccountDB()
}

func createDB(){
	_sql :=`
	CREATE TABLE entry (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	title	TEXT NOT NULL,
	url	TEXT,
	baseTime	INTEGER,
	beginTime	INTEGER NOT NULL,
	endTime		INTEGER NOT NULL,
	site		INTEGER NOT NULL
	);
	
	CREATE TABLE entryToTag (
	id	INTEGER NOT NULL UNIQUE,
	entry_id	INTEGER NOT NULL,
	tag_id	INTEGER NOT NULL,
	PRIMARY KEY(id,entry_id,tag_id)
	);

	CREATE TABLE tag (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	name	TEXT NOT NULL,
	count	INTEGER NOT NULL DEFAULT 0,
	type	INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE sitePage (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	addr	TEXT NOT NULL,
	name	TEXT,
	page	TEXT NOT NULL,
	pcount	INTEGER NOT NULL,
	raw	TEXT NOT NULL,
	dateText	TEXT NOT NULL,
	dateFormat	TEXT,
	urlText		TEXT NOT NULL,
	titleText	TEXT NOT NULL,
	entryId		INTEGER
	);

	`
	HandDB(Conf.DbPath,func(db *sql.DB){
		_,err := db.Exec(_sql)
		if err != nil {
			panic(err)
		}
	})

}
func ClientDo(path string, hand func(io.Reader)error) error {


	Req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return err
	}
	Req.Header = Conf.Header
	var Client http.Client
	res, err := Client.Do(Req)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Second*5)
		return ClientDo(path,hand)
	}
	if res.StatusCode != 200 {
		var da [1024]byte
		n,err := res.Body.Read(da[0:])
		res.Body.Close()
		return fmt.Errorf("status code %d %s %s", res.StatusCode, path,string(da[:n]),err)
	}
	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
		//defer reader.Close()
	default:
		reader = res.Body
	}
	if hand != nil {
		err = hand(reader)
	}
	reader.Close()
	return err

}
func HandDB(dbfile string,handle func(*sql.DB)){

	DB,err := sql.Open(DBType,dbfile)
	if err != nil {
		panic(err)
	}
	handle(DB)
	DB.Close()

}

func HandDBForBack(dbfile string,handle func(*sql.DB) error) (err error){

	DB,err := sql.Open(DBType,dbfile)
	if err != nil {
		return err
	}
	err = handle(DB)
	DB.Close()
	return

}
func Collection(site *SitePage,hand func(en *Entry) error){
	var err error
	var tmpEntrys []*Entry
	//for i:= 1;i<=site.PCount;{
	for i:= 1;i<=1;{
		err  = ClientDo(site.GetUrl(i),func(body io.Reader) error {
			return site.ReadPage(body,func(en *Entry)error{
				tmpEntrys = append(tmpEntrys,en)
				return hand(en)
			})
		})
		if err != nil {
			if err == io.EOF{
				break
			}
			log.Println(err)
		}else{
			i++
		}
	}

}
func runColl () {
	for  {
		<-time.Tick(time.Hour*24)
		log.Println("run coll")
		for _,si := range SiteRun {
			go si.Collection()
		}

	}
}
