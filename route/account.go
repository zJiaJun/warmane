package route

import (
	"github.com/gin-gonic/gin"
	"github.com/zJiajun/warmane/constant"
	"github.com/zJiajun/warmane/engine"
	"github.com/zJiajun/warmane/model/table"
	"net/http"
	"strconv"
)

func listAccountHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		accounts, err := engine.ListAccount()
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withHTMLData(c, "account", accounts)
	}
}

func createAccountHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var account table.Account
		if err := c.ShouldBindJSON(&account); err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		account.Host = constant.HOST
		account.Status = constant.OFFLINE_STATUS
		rows, err := engine.CreateAccount(&account)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withJSONData(c, rows, "Account created successfully")
	}
}

func updateAccountHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		var account table.Account
		if err := c.ShouldBindJSON(&account); err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		account.ID = uint(id)
		rows, err := engine.UpdateAccount(&account)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withJSONData(c, rows, "Account updated successfully")
	}
}

func deleteAccountHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		rows, err := engine.DeleteAccount(id)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withJSONData(c, rows, "Account deleted successfully")
	}
}

func checkAccountHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		result, err := engine.CheckAccount(id)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withJSONData(c, result, "Account checked successfully")
	}
}
