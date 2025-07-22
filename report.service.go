package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReportService struct{}

// const defaultDir = "C:\\Users\\Jiayi Xu\\Desktop\\重要文件\\研报"

func NewReportService() *ReportService {
	return &ReportService{}
}

func (rSvc *ReportService) InitReadStatusOfAllReports(rootPath string, ctx *gin.Context) ([]*GeneralizedReport, error) {
	// step 1: 	遍历 rootPath 构建目录结构树
	rootFolder, err := genDirTreeByReport(rootPath)
	if err != nil {
		return nil, err
	}

	// step 2: 	后序遍历目录结构树计算文件数目/已读数目
	//			输入目录树的根节点, 返回 list, list-item 应包含各目录节点信息, 以及所有子 pdf 数, 所有子 pdf 的已读/已阅数目
	//			前端需要的字段有:	1. key - pathname 	2. childrenKeys		3. readPdfs		4. totalPdfs
	stack := []*MultiWayDirNode{rootFolder}
	var stackTop *MultiWayDirNode

	parentPath := rootPath
	currPath := ""
	reportMap := make(map[string]*GeneralizedReport, 0)
	reportMap[parentPath] = &GeneralizedReport{
		Name:                 parentPath,
		IsDir:                true,
		IsFile:               false,
		IsRead:               false,
		Ext:                  "",
		IsPdf:                false,
		PdfReadCnt:           0,
		FileCnt:              0,
		PdfCnt:               0,
		SelfPath:             parentPath,
		ParentPath:           "",
		Level:                0,
		HasIndustrialReports: false,
	}

	for len(stack) > 0 {
		stackTop = stack[len(stack)-1]
		parentPath = stackTop.Val.Path

		if len(stackTop.Children) == 0 || stackTop.Visited == len(stackTop.Children)-1 {
			//	1. 	叶子节点场景
			//	或
			// 	2.	其所有子树均已处理的场景
			topEntry := stackTop.Val.Entry
			if topEntry == nil {
				// 根节点的情形, Entry == nil
				// 此时要把所有子孙节点的文件计数加上来
				report4Entry := reportMap[parentPath]
				childrenEntries := stackTop.Children
				fileCnt := 0
				pdfCnt := 0
				pdfReadCnt := 0
				hasIndustrialReports := report4Entry.HasIndustrialReports
				for _, childEntry := range childrenEntries {
					currChildEntry := reportMap[childEntry.Val.Path]
					fileCnt += currChildEntry.FileCnt
					pdfCnt += currChildEntry.PdfCnt
					pdfReadCnt += currChildEntry.PdfReadCnt
					if currChildEntry.HasIndustrialReports {
						hasIndustrialReports = true
					}
				}
				report4Entry.FileCnt = fileCnt
				report4Entry.PdfCnt = pdfCnt
				report4Entry.PdfReadCnt = pdfReadCnt
				report4Entry.HasIndustrialReports = hasIndustrialReports
				stack = stack[:len(stack)-1]
				continue
			}

			topEntryName := topEntry.Name()
			report4Entry := reportMap[parentPath]
			if topEntry.Type().IsRegular() {
				// case 1: 	是文件
				report4Entry.FileCnt = 1
				// case 2: 	而且是 pdf 文件
				if path.Ext(topEntryName) == ".pdf" {
					report4Entry.PdfCnt = 1
					if strings.Contains(topEntryName, "已读") || strings.Contains(topEntryName, "已阅") {
						report4Entry.IsRead = true
						report4Entry.PdfReadCnt = 1
					}
				}
			} else if topEntry.Type().IsDir() {
				// case 3: 	是目录
				childrenEntries := stackTop.Children
				fileCnt := 0
				pdfCnt := 0
				pdfReadCnt := 0
				hasIndustrialReports := report4Entry.HasIndustrialReports
				for _, childEntry := range childrenEntries {
					currChildEntry := reportMap[childEntry.Val.Path]
					fileCnt += currChildEntry.FileCnt
					pdfCnt += currChildEntry.PdfCnt
					pdfReadCnt += currChildEntry.PdfReadCnt
					if currChildEntry.HasIndustrialReports {
						hasIndustrialReports = true
					}
				}
				report4Entry.FileCnt = fileCnt
				report4Entry.PdfCnt = pdfCnt
				report4Entry.PdfReadCnt = pdfReadCnt
				report4Entry.HasIndustrialReports = hasIndustrialReports
			} else {
				fmt.Printf("either IsDir nor IsRegular, current entry type:  %s\n", topEntry.Type())
			}

			// 出栈
			stack = stack[:len(stack)-1]
			continue
		}

		stackTop.Visited++
		currChildren := stackTop.Children
		nextChild := currChildren[stackTop.Visited]
		nextEntry := nextChild.Val.Entry
		entryName := nextEntry.Name()
		entryExt := path.Ext(entryName)
		currPath = fmt.Sprintf("%s\\%s", parentPath, entryName)
		isFile := nextEntry.Type().IsRegular()

		reportMap[currPath] = &GeneralizedReport{
			Name:                 entryName,
			IsDir:                nextEntry.Type().IsDir(),
			IsFile:               isFile,
			IsRead:               false,
			Ext:                  entryExt,
			IsPdf:                isFile && entryExt == ".pdf",
			PdfCnt:               0,
			PdfReadCnt:           0,
			FileCnt:              0,
			SelfPath:             currPath,
			Level:                reportMap[parentPath].Level + 1,
			ParentPath:           parentPath,
			HasIndustrialReports: entryName == "行业研报",
		}

		stack = append(stack, nextChild)
	}

	resReportList := []*GeneralizedReport{}
	for _, report := range reportMap {
		resReportList = append(resReportList, report)
	}
	// 	step 3: 	数据写入 db.
	//	TODO:	step 3
	return resReportList, nil
}

func genDirTreeByReport(rootPath string) (*MultiWayDirNode, error) {
	defaultRoot := "C:\\Users\\Jiayi Xu\\Desktop\\研报"
	if rootPath == "" {
		rootPath = defaultRoot
	}

	rootEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	rootChildren := []*MultiWayDirNode{}
	for _, rEntry := range rootEntries {
		rootChildren = append(rootChildren, &MultiWayDirNode{
			Children: []*MultiWayDirNode{},
			Visited:  -1,
			Val: &DirVal{
				Entry: rEntry,
				Path:  fmt.Sprintf("%s\\%s", rootPath, rEntry.Name()),
			},
		})
	}
	rootFolder := &MultiWayDirNode{
		Children: rootChildren,
		Visited:  -1,
		Val: &DirVal{
			Entry: nil,
			Path:  rootPath,
		},
	}

	// 通过广度优先搜索, 构建以 rootFolder 为根节点的完整目录树结构
	q := []*MultiWayDirNode{rootFolder}
	for len(q) > 0 {
		curr := q[0]
		for _, currChild := range curr.Children {
			childVal := currChild.Val
			if childVal.Entry.IsDir() {
				childPath := childVal.Path
				lv2Entries, err := os.ReadDir(childPath)
				if err != nil {
					return nil, err
				}
				lv2Children := []*MultiWayDirNode{}
				for _, lv2Entry := range lv2Entries {
					lv2Children = append(lv2Children, &MultiWayDirNode{
						Children: []*MultiWayDirNode{},
						Visited:  -1,
						Val: &DirVal{
							Path:  fmt.Sprintf("%s\\%s", childPath, lv2Entry.Name()),
							Entry: lv2Entry,
						},
					})
				}
				currChild.Children = lv2Children
				q = append(q, &MultiWayDirNode{
					Children: lv2Children,
					Visited:  -1,
					Val: &DirVal{
						Path:  childPath,
						Entry: childVal.Entry,
					},
				})
			}
		}
		q = q[1:]
	}

	return rootFolder, nil
}

func (rSvc *ReportService) InitReadStatusOfAllReportsV2(rootPath string, ctx *gin.Context) (interface{}, error) {
	// 直接遍历 "C:\\Users\\Jiayi Xu\\Desktop\\研报", 并生成所有文件夹下 pdf 记数的结果。跳过目录结构树这步
	return nil, nil
}
