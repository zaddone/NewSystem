package server
import(
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"database/sql"
	"strings"
	"log"
	"fmt"
)

func loadRouter(){
	Router = gin.Default()
	Router.Static("/static","./static")
	Router.LoadHTMLGlob(Conf.Templates)
	Router.GET("/",func(c *gin.Context){
		c.HTML(http.StatusOK,
		"index.tmpl",
		(func() (sites []*SitePage) {
			ReadSitePage(func(s *SitePage){
				//log.Println(s.Id,s.Name)
				sites = append(sites,s)
			})
			return
		}()))
	})
	Router.POST("/form_post",func(c *gin.Context){
		pcount,err :=strconv.Atoi(c.PostForm("pcount"))
		if err != nil {
			//log.Println(err)
			c.JSON(http.StatusNotFound,err.Error())
			return
		}
		si := NewSitePage(
			c.PostForm("name"),
			c.PostForm("addr"),
			c.PostForm("page"),
			c.PostForm("raw"),
			c.PostForm("date_text"),
			c.PostForm("date_format"),
			c.PostForm("url_text"),
			c.PostForm("title_text"),
			pcount)
		var id int = 0
		id,_ = strconv.Atoi(c.DefaultPostForm("id","0"))
		si.Id = int64(id)
		err = HandDBForBack(Conf.DbPath,func(db *sql.DB)error{
			return si.Save(db)
		})
		if err != nil {
			c.JSON(http.StatusNotFound,err.Error())
			return
		}
		c.JSON(http.StatusOK,si)

	})
	Router.GET("/sites",func(c *gin.Context){
		switch c.Query("content_type"){
		case "json":
			c.JSON(http.StatusOK,
			(func() (sites []*SitePage) {
				ReadSitePage(func(s *SitePage){
					//log.Println(s.Id,s.Name)
					sites = append(sites,s)
				})
				return
			}()))

		default:
			c.HTML(http.StatusOK,
			"sites.tmpl",
			(func() (sites []*SitePage) {
				ReadSitePage(func(s *SitePage){
					//log.Println(s.Id,s.Name)
					sites = append(sites,s)
				})
				return
			}()))
		}
	})
	Router.POST("/run/:id",func(c *gin.Context){
		id,err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound,err.Error())
			return
		}

		si := SiteRun[int64(id)]
		if si != nil {
			c.JSON(http.StatusOK,"is run")
			return
		}
		si,err = FindSitePage(id)
		if err != nil {
			c.JSON(http.StatusNotFound,err.Error())
			return
		}
		SiteRun[si.Id] = si
		go si.Collection()
		c.JSON(http.StatusOK, "Now run")
		return
	})
	Router.GET("/down",func(c *gin.Context){
		c.Redirect(http.StatusMovedPermanently, "/static/js/bootstrap-select.js.map")
	})
	Router.GET("/show",func(c *gin.Context){
		//log.Println("contentType",c.Request.Header.Get("Accept"))
		if strings.Contains(c.Request.Header.Get("Accept"),"json"){
			var where []string
			var val []interface{}

			var w []string
			for _,_s := range c.QueryArray("site"){
				_i,err := strconv.Atoi(_s)
				if err != nil {
					c.JSON(http.StatusNotFound,err.Error())
				}
				val = append(val,_i)
				w = append(w,"?")
			}
			if len(w) > 0 {
				where = append(where,fmt.Sprintf("site in (%s)",strings.Join(w,",")))
				//val = append(val,site)
			}
			ens,err:=GetEntryArr(c,where,val)
			if err != nil {
				c.JSON(http.StatusNotFound,err.Error())
				return
			}
			c.JSON(http.StatusOK, gin.H{"ens":ens})
			return

		}else{
			c.JSON(http.StatusNotFound,nil)
			//c.HTML(http.StatusOK,"show.tmpl",nil)
			return
		}
	})
	Router.Run(Conf.Port)
}
func GetEntryArr(c *gin.Context,where []string,val []interface{}) (ens []*Entry,err error) {
	offset,err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		return nil,err
	}
	limit,err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		return nil,err
	}
	begin,err := strconv.Atoi(c.DefaultQuery("begin","0"))
	if err != nil {
		return nil,err
	}
	end,err := strconv.Atoi(c.DefaultQuery("end","0"))
	if err != nil {
		return nil,err
	}
	if begin > 0 {
		where = append(where,"baseTime >= ?")
		val = append(val,begin)
	}
	if end  > 0 {
		where = append(where,"baseTime <= ?")
		val = append(val,end)
	}
	val = append(val,limit,offset)
	var sql_  string
	if len(where)>0 {
		sql_ = " WHERE " + strings.Join(where," and ")
	}
	sql_ += " LIMIT ? OFFSET ?"
	log.Println(sql_,val)
	err = ReadEntry(func(en *Entry)error{
		ens = append(ens,en)
		return nil
	},sql_,val...)
	return

}
