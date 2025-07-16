package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	ReportSvc *ReportService
}

// const defaultDir = "C:\\Users\\Jiayi Xu\\Desktop\\重要文件\\研报"

func NewReportHanlder() *ReportHandler {
	return &ReportHandler{
		ReportSvc: NewReportService(),
	}
}

func (hdl *ReportHandler) RegisterRoutes(s *gin.Engine) {
	reportG := s.Group("/report")
	reportG.GET("/initall", hdl.InitReadStatusOfAllReports)
	reportG.GET("/initall/v2", hdl.InitReadStatusOfAllReportsV2)
	reportG.GET("/viewdir", hdl.ViewDirAndFilesByPath)
	reportG.GET("/testlogic", hdl.TestLogic)
}

func (hdl *ReportHandler) InitReadStatusOfAllReports(ctx *gin.Context) {
	var resp WrapperResp
	// rootPath := "C:\\"
	rootPath := "C:\\Users\\Jiayi Xu\\Desktop\\研报"
	reportsList, err := hdl.ReportSvc.InitReadStatusOfAllReports(rootPath, ctx)
	if err != nil {
		resp = WrapperResp{
			Data: nil,
			Msg:  fmt.Sprintf("%+v", err),
			Code: 1,
		}
		ctx.JSON(400, resp)
		return
	}

	resp = WrapperResp{
		Data: struct {
			Reports []*GeneralizedReport `json:"reports"`
		}{
			Reports: reportsList,
		},
		Code: 0,
		Msg:  "Init read status of all reports succeeds !",
	}
	ctx.JSON(200, resp)
}

func (hdl *ReportHandler) InitReadStatusOfAllReportsV2(ctx *gin.Context) {
	rootPath := "C:\\Users\\Jiayi Xu\\Desktop\\研报"
	_, err := hdl.ReportSvc.InitReadStatusOfAllReportsV2(rootPath, ctx)
	if err != nil {
		// TODO:	remove this line
		fmt.Println(err)
		return
	}
	resp := WrapperResp{
		Data: struct{}{},
		Code: 0,
		Msg:  "Reset read status of all reports succeeds !",
	}
	ctx.JSON(200, resp)
}

// view files and dirs in by specific path
// TODO: should read data from db
func (hdl *ReportHandler) ViewDirAndFilesByPath(ctx *gin.Context) {
	var resp WrapperResp
	parentPath, found := ctx.GetQuery("rootpath")
	if !found {
		parentPath = "C:\\"
	}

	entries, err := os.ReadDir(parentPath)
	if err != nil {
		resp = WrapperResp{
			Code: 1,
			Data: struct{}{},
			Msg:  fmt.Sprintf("folder %s NOT exist", parentPath),
		}
		ctx.JSON(400, resp)
		return
	}

	reportsInfo := []GeneralizedReport{}
	var isDir bool
	var entryName string
	for _, entry := range entries {
		isDir = entry.Type().IsDir()
		entryName = entry.Name()
		entryExt := path.Ext(entryName)
		reportsInfo = append(reportsInfo, GeneralizedReport{
			// Name:     entryName,
			IsDir:    isDir,
			IsFile:   entry.Type().IsRegular(),
			IsRead:   isDir && (strings.Contains(entryName, "已阅") || strings.Contains(entryName, "已读")),
			Ext:      entryExt,
			IsPdf:    entryExt == ".pdf",
			SelfPath: fmt.Sprintf("%s\\%s", parentPath, entryName),
		})
	}

	resp = WrapperResp{
		Code: 0,
		Data: struct {
			FileNames []GeneralizedReport `json:"fileNames"`
		}{
			FileNames: reportsInfo,
		},
		Msg: "pull all reports succeeds",
	}
	ctx.JSON(200, resp)
}

func (hdl *ReportHandler) TestLogic(ctx *gin.Context) {
	// var resp WrapperResp
	// queryPath, found := ctx.GetQuery("rootpath")
	// if !found {
	// 	ctx.JSON(404, "Not Found")
	// 	return
	// }
	// entries, err := os.ReadDir(queryPath)
	resp := WrapperResp{
		Code: 0,
		Data: struct {
			FileNames string `json:"fileNames"`
		}{
			FileNames: "",
		},
		Msg: "Test Done",
	}
	ctx.JSON(200, resp)
}
