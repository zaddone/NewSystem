package server
import(
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"database/sql"
	"strings"
	//"log"
)

func loadRouter(){
	Router = gin.Default()
	Router.LoadHTMLGlob(Conf.Templates)
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
	Router.GET("/show/:id",func(c *gin.Context){
		id,err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound,err.Error())
			return
		}
		switch c.Query("content_type"){
		case "json":
			var where []string
			var val []interface{}
			where = append(where,"site = ?")
			val = append(val,id)
			ens,err:=GetEntryArr(c,where,val)
			if err != nil {
				c.JSON(http.StatusNotFound,err.Error())
				return
			}
			c.JSON(http.StatusOK, gin.H{"ens":ens})
			return
		default:
			c.HTML(http.StatusOK,
			"show.tmpl",
			gin.H{"site":id})
			return
		}

	})
	Router.GET("/show",func(c *gin.Context){
		switch c.Query("content_type"){
		case "json":
			var where []string
			var val []interface{}
			site,err := strconv.Atoi(c.Query("site"))
			if (err == nil) && (site > 0) {
				where = append(where,"site = ?")
				val = append(val,site)
			}
			ens,err:=GetEntryArr(c,where,val)
			if err != nil {
				c.JSON(http.StatusNotFound,err.Error())
				return
			}
			c.JSON(http.StatusOK, gin.H{"ens":ens})
			return

		default:
			c.HTML(http.StatusOK,
			"show.tmpl",
			gin.H{"site":0})
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
	begin,err := strconv.Atoi(c.DefaultQuery("base_begin","0"))
	if err != nil {
		return nil,err
	}
	end,err := strconv.Atoi(c.DefaultQuery("base_end","0"))
	if err != nil {
		return nil,err
	}
	if begin > 0 {
		where = append(where,"baseTime >= ?")
		val = append(val,begin)
	}
	if end  > 0 {
		where = append(where,"baseTime <= ?")
		val = append(val,begin)
	}
	val = append(val,limit,offset)
	var sql_  string
	if len(where)>0 {
		sql_ = " WHERE " + strings.Join(where," and ")
	}
	sql_ += " LIMIT ? OFFSET ?"
	err = ReadEntry(func(en *Entry)error{
		ens = append(ens,en)
		return nil
	},sql_,val...)
	return

}
