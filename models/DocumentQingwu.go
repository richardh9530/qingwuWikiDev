package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego"
	"strings"
)

type DocumentQingwuResult struct {
	DocumentId   int    `json:"doc_id"`
	DocumentName string `json:"doc_name"`
	// Identify 文档唯一标识
	Identify     string    `json:"identify"`
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	ModifyTime   time.Time `json:"modify_time"`
	CreateTime   time.Time `json:"create_time"`
	BookId       int       `json:"book_id"`
	BookName     string    `json:"book_name"`
	BookIdentify string    `json:"book_identify"`
	SearchType   string    `json:"search_type"`
}

func RecentDocumentResult() *DocumentQingwuResult {
	return &DocumentQingwuResult{}
}

//分页全局搜索.
func (m *DocumentQingwuResult) FindToPager(keyword string, pageIndex, pageSize, memberId int) (searchResult []*DocumentQingwuResult, totalCount int, err error) {
	o := orm.NewOrm()

	offset := (pageIndex - 1) * pageSize

	keyword = "%" + strings.Replace(keyword," ","%",-1) + "%"

	if memberId <= 0 {
		sql1 := `SELECT count(doc.document_id) as total_count FROM md_documents AS doc
  LEFT JOIN md_books as book ON doc.book_id = book.book_id
WHERE book.privately_owned = 0 AND (doc.document_name LIKE ? OR doc.release LIKE ?) `

		sql2 := `SELECT *
FROM (
       SELECT
         doc.document_id,
         doc.modify_time,
         doc.create_time,
         doc.document_name,
         doc.identify,
         doc.release    AS description,
         book.identify  AS book_identify,
         book.book_name,
         rel.member_id,
         member.account AS author,
         'document'     AS search_type
       FROM md_documents AS doc
         LEFT JOIN md_books AS book ON doc.book_id = book.book_id
         LEFT JOIN md_relationship AS rel ON book.book_id = rel.book_id AND rel.role_id = 0
         LEFT JOIN md_members AS member ON rel.member_id = member.member_id
       WHERE book.privately_owned = 0 AND (doc.document_name LIKE ? OR doc.release LIKE ?)
       UNION ALL
       SELECT
         blog.blog_id,
         blog.modify_time,
         blog.create_time,
         blog.blog_title,
         blog.blog_identify,
         blog.blog_release,
         blog.blog_identify,
         blog.blog_title,
         blog.member_id,
         member.account,
         'blog' AS search_type
       FROM md_blogs AS blog
         LEFT JOIN md_members AS member ON blog.member_id = member.member_id
       WHERE blog.blog_status = 'public' AND (blog.blog_release LIKE ? OR blog.blog_title LIKE ?)
     ) AS union_table
ORDER BY modify_time DESC
LIMIT ?, ?;`

		err = o.Raw(sql1, keyword, keyword).QueryRow(&totalCount)
		if err != nil {
			beego.Error("查询搜索结果失败 -> ",err)
			return
		}
		sql3 := `       SELECT
         count(*)
       FROM md_blogs AS blog
       WHERE blog.blog_status = 'public' AND (blog.blog_release LIKE ? OR blog.blog_title LIKE ?);`

		c := 0
		err = o.Raw(sql3, keyword, keyword).QueryRow(&c)
		if err != nil {
			beego.Error("查询搜索结果失败 -> ",err)
			return
		}

		totalCount += c

		_, err = o.Raw(sql2, keyword, keyword,keyword,keyword, offset, pageSize).QueryRows(&searchResult)
		if err != nil {
			beego.Error("查询搜索结果失败 -> ",err)
			return
		}
	} else {
		sql1 := `SELECT count(doc.document_id) as total_count FROM md_documents AS doc
  LEFT JOIN md_books as book ON doc.book_id = book.book_id
  LEFT JOIN md_relationship AS rel ON doc.book_id = rel.book_id AND rel.role_id = 0
  LEFT JOIN md_relationship AS rel1 ON doc.book_id = rel1.book_id AND rel1.member_id = ?
			left join (select * from (select book_id,team_member_id,role_id
                   	from md_team_relationship as mtr
					left join md_team_member as mtm on mtm.team_id=mtr.team_id and mtm.member_id=? order by role_id desc )as t group by t.role_id,t.team_member_id,t.book_id) as team 
					on team.book_id = book.book_id
WHERE (book.privately_owned = 0 OR rel1.relationship_id > 0 or team.team_member_id > 0)  AND (doc.document_name LIKE ? OR doc.release LIKE ?) `

		sql2 := `SELECT *
FROM (
       SELECT
         doc.document_id,
         doc.modify_time,
         doc.create_time,
         doc.document_name,
         doc.identify,
         doc.release    AS description,
         book.identify  AS book_identify,
         book.book_name,
         rel.member_id,
         member.account AS author,
         'document'     AS search_type
       FROM md_documents AS doc
         LEFT JOIN md_books AS book ON doc.book_id = book.book_id
         LEFT JOIN md_relationship AS rel ON book.book_id = rel.book_id AND rel.role_id = 0
         LEFT JOIN md_members AS member ON rel.member_id = member.member_id
         LEFT JOIN md_relationship AS rel1 ON doc.book_id = rel1.book_id AND rel1.member_id = ?
         LEFT JOIN (SELECT *
                    FROM (SELECT
                            book_id,
                            team_member_id,
                            role_id
                          FROM md_team_relationship AS mtr
                            LEFT JOIN md_team_member AS mtm ON mtm.team_id = mtr.team_id AND mtm.member_id = ?
                          ORDER BY role_id DESC) AS t
                    GROUP BY t.role_id, t.team_member_id, t.book_id) AS team
           ON team.book_id = book.book_id
       WHERE (book.privately_owned = 0 OR rel1.relationship_id > 0 OR team.team_member_id > 0) AND
             (doc.document_name LIKE ? OR doc.release LIKE ?)
       UNION ALL

       SELECT
         blog.blog_id,
         blog.modify_time,
         blog.create_time,
         blog.blog_title,
         blog.blog_identify,
         blog.blog_release,
         blog.blog_identify,
         blog.blog_title,
         blog.member_id,
         member.account,
         'blog' AS search_type
       FROM md_blogs AS blog
         LEFT JOIN md_members AS member ON blog.member_id = member.member_id
       WHERE (blog.blog_status = 'public' OR blog.member_id = ?) AND blog.blog_type = 0 AND
             (blog.blog_release LIKE ? OR blog.blog_title LIKE ?)
     ) AS union_table
ORDER BY modify_time DESC
LIMIT ?, ?;`

		err = o.Raw(sql1, memberId, memberId, keyword, keyword).QueryRow(&totalCount)
		if err != nil {
			return
		}
		sql3 := `       SELECT
         count(*)
       FROM md_blogs AS blog
       WHERE (blog.blog_status = 'public' OR blog.member_id = ?) AND blog.blog_type = 0 AND
             (blog.blog_release LIKE ? OR blog.blog_title LIKE ?);`

		c := 0
		err = o.Raw(sql3,memberId, keyword, keyword).QueryRow(&c)
		if err != nil {
			beego.Error("查询搜索结果失败 -> ",err)
			return
		}

		totalCount += c
		_, err = o.Raw(sql2, memberId, memberId, keyword, keyword,memberId,keyword, keyword, offset, pageSize).QueryRows(&searchResult)
		if err != nil {
			return
		}
	}
	return
}

//项目内搜索.
func (m *DocumentQingwuResult) SearchDocument(keyword string, book_id int) (docs []*DocumentQingwuResult, err error) {
	o := orm.NewOrm()

	sql := "SELECT * FROM md_documents WHERE book_id = ? AND (document_name LIKE ? OR `release` LIKE ?) "
	keyword = "%" + keyword + "%"

	_, err = o.Raw(sql, book_id, keyword, keyword).QueryRows(&docs)

	return
}
