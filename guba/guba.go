package pholcus_lib

// 基础包
import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	"github.com/henrylee2cn/pholcus/common/goquery"
	"github.com/henrylee2cn/pholcus/logs"
	//. "github.com/henrylee2cn/pholcus/app/spider/common"    //选用
	//DOM解析
	//信息输出
	// net包
	//设置http.Header
	// "net/url"
	// 编码包
	// "encoding/xml"
	// "encoding/json"
	// 字符串处理包
	// "regexp"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	Guba.Register()
}

type gubaData struct {
	viewCount    string
	commentCount string
	label        string
}

var Guba = &Spider{
	Name:         "股吧热帖",
	Description:  "股吧热帖",
	Keyin:        KEYIN,
	Limit:        LIMIT,
	EnableCookie: false,
	Namespace: func(self *Spider) string {
		return "guba"
	},
	SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
		return "hot"
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:         "http://guba.eastmoney.com/default,99_1.html",
				Rule:        "getFirst",
				ConnTimeout: -1,
				Reloadable:  true,
			})
		},
		Trunk: map[string]*Rule{
			"getOthers": {
				ParseFunc: func(ctx *Context) {
					ctx.GetDom().Find("ul.newlist li").Each(func(i int, s *goquery.Selection) {
						a := s.Children().First()
						viewcount := strings.ReplaceAll(strings.ReplaceAll(a.Text(), " ", ""), "\n", "")
						commentcount := strings.ReplaceAll(strings.ReplaceAll(a.Next().Text(), " ", ""), "\n", "")
						label, _ := a.Next().Next().Find("em").Attr("class")
						url, _ := a.Next().Next().Find("a.note").Attr("href")
						if url[0:2] == "//" {
							// url = "http:" + url
							// ctx.AddQueue(&request.Request{
							// 	Url:         url,
							// 	Rule:        "getAsk",
							// 	ConnTimeout: -1,
							// 	Reloadable:  false,
							// 	AdditionalArgs: gubaData{
							// 		viewCount:    viewcount,
							// 		commentCount: commentcount,
							// 		label:        label,
							// 	},
							// })
						} else {
							url = "http://guba.eastmoney.com" + url
							ctx.AddQueue(&request.Request{
								Url:         url,
								Rule:        "getContent",
								ConnTimeout: -1,
								Reloadable:  true,
								AdditionalArgs: gubaData{
									viewCount:    viewcount,
									commentCount: commentcount,
									label:        label,
								},
							})
						}
					})
				},
			},
			"getFirst": {
				ParseFunc: func(ctx *Context) {
					ctx.GetDom().Find("ul.newlist li").Each(func(i int, s *goquery.Selection) {
						a := s.Children().First()
						viewcount := strings.ReplaceAll(strings.ReplaceAll(a.Text(), " ", ""), "\n", "")
						commentcount := strings.ReplaceAll(strings.ReplaceAll(a.Next().Text(), " ", ""), "\n", "")
						label, _ := a.Next().Next().Find("em").Attr("class")
						url, _ := a.Next().Next().Find("a.note").Attr("href")
						if url[0:2] == "//" {
							url = "http:" + url
							ctx.AddQueue(&request.Request{
								Url:         url,
								Rule:        "getAsk",
								ConnTimeout: -1,
								Reloadable:  true,
								AdditionalArgs: gubaData{
									viewCount:    viewcount,
									commentCount: commentcount,
									label:        label,
								},
							})
						} else {
							url = "http://guba.eastmoney.com" + url
							ctx.AddQueue(&request.Request{
								Url:         url,
								Rule:        "getContent",
								ConnTimeout: -1,
								Reloadable:  true,
								AdditionalArgs: gubaData{
									viewCount:    viewcount,
									commentCount: commentcount,
									label:        label,
								},
							})
						}

					})
					total, err := strconv.Atoi(ctx.GetDom().Find("span.sumpage").Text())
					if err != nil {
						logs.Log.Error(err.Error())
					} else {
						for i := 2; i != total+1; i++ {
							ctx.AddQueue(&request.Request{
								Url:         "http://guba.eastmoney.com/default,99_" + strconv.Itoa(i) + ".html",
								Rule:        "getOthers",
								ConnTimeout: -1,
								Reloadable:  true,
							})
						}
					}

				},
			},
			"getContent": {
				ItemFields: []string{
					"Title",
					"ViewCount",
					"CommentCount",
					"PublishDate",
					"Source",
					"Lable",
					"Content",
					"Author",
					"UUID",
				},
				ParseFunc: func(ctx *Context) {
					title := ctx.GetDom().Find("#zwconttbt").Text()
					publishdate := ctx.GetDom().Find("#zwconttb .zwfbtime").Text()
					publishdate = publishdate[10:29]
					t, _ := time.Parse("2006-01-02 15:04:05", publishdate)
					source := ctx.GetDom().Find("#stockname a").Text()
					tmp := (ctx.GetRequest().GetAdditionalArgs()).(gubaData)
					viewcount := tmp.viewCount
					commentcount := tmp.commentCount
					label := tmp.label
					content := ctx.GetDom().Find("#zwconbody").Text()
					author := ctx.GetDom().Find("#zwconttbn strong a").Text()
					id, _ := uuid.NewUUID()
					ctx.Output(map[int]interface{}{
						0: title,
						1: viewcount,
						2: commentcount,
						3: t.Unix() - 28800,
						4: source,
						5: label,
						6: content,
						7: author,
						8: id.String(),
					})
				},
			},
			"getAsk": {
				ItemFields: []string{
					"Title",
					"ViewCount",
					"CommentCount",
					"PublishDate",
					"Source",
					"Lable",
					"Content",
					"Author",
					"UUID",
				},
				ParseFunc: func(ctx *Context) {
					title := ctx.GetDom().Find("#questionHead .qh_content").Text()
					publishdate := ctx.GetDom().Find("#questionHead .publishTime").Text()
					publishdate = publishdate[9:]
					t, _ := time.Parse("2006-01-02 15:04", publishdate)
					source := "悬赏问答"
					tmp := (ctx.GetRequest().GetAdditionalArgs()).(gubaData)
					label := tmp.label
					viewcount := tmp.viewCount
					commentcount := tmp.commentCount
					content := ctx.GetDom().Find("#questionHead .qh_content").Text()
					author := ctx.GetDom().Find("#questionHead .userName").Text()
					id, _ := uuid.NewUUID()

					ctx.Output(map[int]interface{}{
						0: title,
						1: viewcount,
						2: commentcount,
						3: t.Unix() - 28800,
						4: source,
						5: label,
						6: content,
						7: author,
						8: id.String(),
					})
				},
			},
		},
	},
}
