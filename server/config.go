package server
import(
	"net/http"
	"github.com/BurntSushi/toml"
	"os"
)
type Config struct {
	Proxy string
	Port string
	DbPath string
	Templates string
	Server bool
	Header http.Header
	//Site []*SitePage
}
func (self *Config) Save(){
	fi,err := os.OpenFile(*FileName,os.O_CREATE|os.O_WRONLY,0777)
	//fi,err := os.Open(FileName)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	e := toml.NewEncoder(fi)
	err = e.Encode(self)
	if err != nil {
		panic(err)
	}
}
func NewConfig()  *Config {
	var c Config
	_,err := os.Stat(*FileName)
	if err != nil {
		c.Proxy = ""
		c.Port=":8080"
		c.DbPath = "foo.db"
		c.Templates = "./templates/*"
		c.Server = true
		c.Header = http.Header{
			"Accept":[]string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
			"Connection":[]string{"keep-alive"},
			"Accept-Encoding":[]string{"gzip, deflate, sdch"},
			"Accept-Language":[]string{"zh-CN,zh;q=0.8"},
			"User-Agent":[]string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/58.0.3029.110 Chrome/58.0.3029.110 Safari/537.36"}}
		//c.Site = []*SitePage{
		//	NewSitePage(
		//		"http://www.ccgp-sichuan.gov.cn/CmsNewsController.do?method=recommendBulletinList&moreType=provincebuyBulletinMore&channelCode=sjcg1&rp=25&page=1",
		//		"page",
		//		".list-info .info li",
		//		".time.curr|text",
		//		"022006-01",
		//		"a|href",
		//		"a .title|text",
		//		560)}
		c.Save()
	}else{
		if _,err := toml.DecodeFile(*FileName,&c);err != nil {
			panic(err)
		}
	}
	return &c
}
