package server
import(
	"net/url"
	"fmt"
	"log"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"time"
	"io"
	"regexp"
	"database/sql"
)
var (
	reg *regexp.Regexp = regexp.MustCompile("\\s+")
)

type SitePage struct {
	Id int64
	Addr string  `json:"addr"`
	Name string `json:"name"`
	Url *url.URL
	Page string `json:"page"`
	PCount int `json:"pcount"`
	Raw string `json:"raw"`
	DateText string `json:"dateText"`
	DateFormat string `json:"dateFormat"`
	UrlText string `json:"urlText"`
	TitleText string `json:"titleText"`
	EntryId int64  `json:"entryId"`
	EndEntry *Entry
}
func FindSitePage(id int) (si *SitePage,err error) {

	si = &SitePage{}
	HandDB(Conf.DbPath,func(db *sql.DB){
		err = si.LoadDB(db.QueryRow("SELECT id,name,addr,page,pcount,raw,dateText,dateFormat,urlText,titleText,entryId FROM sitePage WHERE id = ?",id),db)
	})
	return si,err
}
func ReadSitePage(hand func(*SitePage)){
	HandDB(Conf.DbPath,func(db *sql.DB){
		row,err := db.Query("SELECT id,name,addr,page,pcount,raw,dateText,dateFormat,urlText,titleText,entryId FROM sitePage")
		if err != nil {
			panic(err)
		}
		for row.Next() {
			si := &SitePage{}
			err = si.LoadDB(row,db)
			if err != nil {
				panic(err)
			}
			hand(si)
			//go si.Collection()
		}
	})
}
func NewSitePage(name,addr,page,raw,date,format,u,title string,pcount int) *SitePage {

	ur,err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	return &SitePage{
		Name:name,
		Url:ur,
		Addr:addr,
		Page:page,
		DateText:date,
		DateFormat:format,
		UrlText:u,
		TitleText:title,
		PCount:pcount,
		Raw:raw,
		EndEntry:&Entry{}}

}
func (self *SitePage) LoadDB(ro _row,db *sql.DB) (err error) {
	err = ro.Scan(
		&self.Id,
		&self.Name,
		&self.Addr,
		&self.Page,
		&self.PCount,
		&self.Raw,
		&self.DateText,
		&self.DateFormat,
		&self.UrlText,
		&self.TitleText,
		&self.EntryId)
	if err != nil {
		return err
	}
	self.Url,err = url.Parse(self.Addr)
	if err != nil {
		return err
	}
	if self.EntryId >0 {
		self.EndEntry = &Entry{}
		return self.EndEntry.LoadDB(self.EntryId,db)
	}
	return nil

}
func (self *SitePage) GetUrl(page int) string {
	if self.Url == nil {
		var err error
		self.Url,err = url.Parse(self.Addr)
		if err != nil {
			panic(err)
		}
	}
	val := self.Url.Query()
	val.Set(self.Page,fmt.Sprintf("%d",page))
	return self.Url.Scheme+"://"+self.Url.Hostname()+self.Url.EscapedPath()+"?"+val.Encode()
}
func (self *SitePage) Collection() {
	var err error
	var tmpEntrys []*Entry
	//for i:= 1;i<=site.PCount;{
	for i:= 1;i<=1;{
		err = ClientDo(self.GetUrl(i),func(body io.Reader) error {
			return self.ReadPage(body,func(en *Entry) error{
				tmpEntrys = append(tmpEntrys,en)
				return nil
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

	HandDB(Conf.DbPath,func(db *sql.DB){
		for i:= len(tmpEntrys)-1;i>=0;i-- {
			self.EndEntry = tmpEntrys[i]
			res := StructSaveForDB(db,"entry",self.EndEntry)
			self.EndEntry.Id,err = res.LastInsertId()
			if err != nil {
				panic(err)
			}
		}
		self.EntryId = self.EndEntry.Id
		err = self.Save(db)
		if err != nil {
			panic(err)
		}
	})
}
func (self *SitePage) Save(db *sql.DB) (err error) {
	if self.Id <=0 {
		res := StructSaveForDB(db,"sitePage",self)
		self.Id,err = res.LastInsertId()
		return err
	}else{
		res := StructUpdateForDB(db,"sitePage",self,"id",self.Id)
		_,err := res.RowsAffected()
		return err
	}
}

func (self *SitePage) ReadPage(body io.Reader,hand func(en *Entry) error) (err error) {
	doc,err := goquery.NewDocumentFromReader(body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var en *Entry
	doc.Find(self.Raw).EachWithBreak(func(i int, s *goquery.Selection)bool {
		en = &Entry{
			BaseTime : self.Date(s),
			Url : HandQuery(self.UrlText,s),
			Title : HandQuery(self.TitleText,s),
			BeginTime : time.Now().Unix(),
			EndTime : time.Now().Unix(),
			Site:  self.Id}
		if self.EndEntry != nil {
			if self.EndEntry.BaseTime == en.BaseTime {
				if self.EndEntry.Title == en.Title {
					err = io.EOF
					return false
				}
			}
		}
		err = hand(en)
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func (self *SitePage) Date(s *goquery.Selection) int64 {
	d := HandQuery(self.DateText,s)
	//reg := regexp.MustCompile("\\s+")
	//d = reg.ReplaceAllString(d,"")
	if d == "" {
		return 0
	}
	ti , err := time.Parse(self.DateFormat,d)
	if err != nil {
		//fmt.Println(d)
		panic(err)
	}
	return ti.Unix()
}
func HandQuery(v string,s *goquery.Selection) (_v string) {
	vs := strings.Split(v,"|")
	if len(vs) <2 {
		return ""
	}
	switch vs[1]  {
	case "text":
		_v = s.Find(vs[0]).Text()
	default:
		_v,_ = s.Find(vs[0]).Attr(vs[1])
	}
	return reg.ReplaceAllString(_v,"")
}
