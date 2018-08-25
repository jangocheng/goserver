package webs

import (
	"github.com/cihub/seelog"
	"strings"
	"github.com/gin-gonic/gin"
	"html/template"
	"utils/gpa"
	"strconv"
	"github.com/gin-contrib/multitemplate"
	"path/filepath"
)

var FunConstantMaps = template.FuncMap{
	"unescaped": func(x string) template.HTML { return template.HTML(x) },
}

// WebTplRenderCreate("ui/views/wk-site/", "/**/*", "/*")
func WebTplRenderCreate(templatesDir string, pages ...string) multitemplate.Renderer {
	ext := ".html"
	r := multitemplate.NewRenderer()
	webMs, _ := filepath.Glob(templatesDir + "layout/web/*" + ext)
	h5Ms, _ := filepath.Glob(templatesDir + "layout/h5/*" + ext)
	seelog.Info("web modules:", webMs)
	for _, p := range pages {
		pages, err := filepath.Glob(templatesDir + p + ext)
		if err != nil {
			panic(err.Error())
		}
		for _, page := range pages {
			j := strings.LastIndex(page, "-")
			if j < 0 || j < len(templatesDir) {
				if strings.Index(page, "layout") < 0 {
					seelog.Warn("命名规则不对:-ua[].html")
				}
				continue
			}
			layout := templatesDir + "layout/" + page[j+1:]
			n := page[0:len(page)-len(ext)][len(templatesDir):]
			d := strings.Index(n, ",")
			if d > 0 {
				n = n[0:d]
			}
			n = strings.Replace(n, "\\", "/", -1)
			ms := []string{layout, page}
			if strings.Index(n, "-web") > 0 {
				if len(webMs) > 0 {
					ms = append(ms, webMs...)
				}
			} else {
				if len(h5Ms) > 0 {
					ms = append(ms, h5Ms...)
				}
			}
			seelog.Info(n, ms)
			//r.AddFromFiles(n, FunConstantMaps, ms...)

			r.AddFromFilesFuncs(n, FunConstantMaps, ms...)
		}
	}
	return r
}

type WebTemplate struct {
	*WebBase
	TplName string //模版名称
}

func (p *WebTemplate) GetTplName() string {
	if len(p.TplName) > 1 {
		return p.TplName
	}
	url := p.Context.Request.URL.Path
	if len(url) == 1 {
		url = "/index"
	}
	return url[1:]
}

func WebTplWithSp(ctx *gin.Context, g *gpa.Gpa, auth func(c *gin.Context) (bool, int64)) {
	if strings.Index(ctx.Request.URL.Path, ".") > 0 {
		seelog.Warn("404:", ctx.Request.URL.Path)
		ctx.AbortWithStatus(404)
		return
	}
	tpl := &WebTemplate{}
	tpl.WebBase = WebBaseNew(ctx)
	tplName := tpl.GetTplName()
	ns := strings.Split(tplName, "/")
	spName := "p"
	for _, n := range ns {
		if len(n) > 1 {
			spName += strings.ToUpper(n[0:1]) + n[1:]
		}
	}
	code := SpExec(spName, g, tpl.WebBase, auth)
	if code == 200 {
		ctx.HTML(200, tplName+"-"+tpl.Ua, tpl.Out)
	} else {
		seelog.Error("code=", code, ",spName=", spName, ",tplName=", tplName)
		ctx.HTML(200, strconv.Itoa(code)+"-"+tpl.Ua, tpl)
	}
}
