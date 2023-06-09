package v1

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/abdulkarimogaji/blognado/db"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-sql-driver/mysql"
)

func createBlog(dbService db.DBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body db.CreateBlogRequest
		err := c.ShouldBindBodyWith(&body, binding.JSON)
		if err != nil {
			validationResponse(err, c)
			return
		}

		blog, err := dbService.CreateBlog(c, body)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "user does not exist",
					"error":   err.Error(),
					"data":    nil,
				})
				return
			}

			me, ok := err.(*mysql.MySQLError)
			if ok && me.Number == MYSQL_KEY_EXISTS {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "blog slug already exists",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "server error",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Blog created successfully",
			"data":    blog,
		})
	}
}

func getBlogByIdOrSlug(dbService db.DBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idOrSlug := c.Param("idOrSlug")
		var blog db.Blog
		id, err := strconv.Atoi(idOrSlug)
		if err == nil {
			blog, err = dbService.GetBlogById(c, id)
		} else {
			blog, err = dbService.GetBlogBySlug(c, idOrSlug)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "blog not found",
					"error":   err.Error(),
					"data":    nil,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "server error",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "success",
			"data":    blog,
		})
	}
}

type getBlogQuery struct {
	Page         string `form:"page"`
	Limit        string `form:"limit"`
	Title        string `form:"title"`
	AuthorId     string `form:"author_id"`
	AuthorName   string `form:"author_name"`
	PostedAfter  string `form:"posted_after"`
	PostedBefore string `form:"posted_before"`
}

func getBlogsPaginate(dbService db.DBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query getBlogQuery
		c.ShouldBindQuery(&query)
		filters, paginationParams := parseBlogQueryParams(query)

		blogs, total, err := dbService.GetBlogs(c, filters, paginationParams)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "server error",
				"error":   err.Error(),
			})
			return
		}

		var totalPages, currentPage, pageSize *int

		if paginationParams.Limit == 0 {
			totalPages = nil
			currentPage = nil
			pageSize = nil
		} else {
			tmp := total / paginationParams.Limit
			tmp2 := 1
			totalPages = &tmp
			if paginationParams.Page > 0 {
				tmp2 = paginationParams.Page
			}
			currentPage = &tmp2
			pageSize = &paginationParams.Limit
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Blog created successfully",
			"data": gin.H{
				"list":        blogs,
				"total":       total,
				"page":        currentPage,
				"page_size":   pageSize,
				"total_pages": totalPages,
			},
		})

	}
}

func parseBlogQueryParams(query getBlogQuery) (db.GetBlogsFilters, db.PaginationParams) {
	var limit, page int
	limit, _ = strconv.Atoi(query.Limit)
	page, _ = strconv.Atoi(query.Page)

	// get filters
	var title, postedBefore, postedAfter, authorName *string
	var authorId *int

	a_id, _ := strconv.Atoi(query.AuthorId)

	if a_id == 0 {
		authorId = nil
	} else {
		authorId = &a_id
	}

	if query.Title == "" {
		title = nil
	} else {
		title = &query.Title
	}

	if query.AuthorName == "" {
		authorName = nil
	} else {
		authorName = &query.AuthorName
	}

	if _, err := time.Parse(time.DateTime, query.PostedAfter); err != nil || query.PostedAfter == "" {
		postedAfter = nil
	} else {
		postedAfter = &query.PostedAfter
	}

	if _, err := time.Parse(time.DateTime, query.PostedBefore); err != nil || query.PostedBefore == "" {
		postedBefore = nil
	} else {
		postedBefore = &query.PostedBefore
	}

	return db.GetBlogsFilters{Title: title,
			PostedAfter:  postedAfter,
			PostedBefore: postedBefore,
			AuthorId:     authorId, AuthorName: authorName}, db.PaginationParams{Page: page,
			Limit: limit}
}
