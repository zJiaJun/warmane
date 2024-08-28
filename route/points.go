package route

import (
	"github.com/gin-gonic/gin"
	"github.com/zJiajun/warmane/common"
	"github.com/zJiajun/warmane/engine"
	"github.com/zJiajun/warmane/model/table"
	"net/http"
	"strconv"
)

func listAccountDetailsHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountDetails, err := engine.ListAccountDetails()
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		var pairs []common.Pair[*table.AccountDetails, *table.AccountDetails]
		for i := 0; i < len(accountDetails); i += 2 {
			if i+1 < len(accountDetails) {
				pairs = append(pairs, common.Pair[*table.AccountDetails, *table.AccountDetails]{Left: accountDetails[i], Right: accountDetails[i+1]})
			} else {
				pairs = append(pairs, common.Pair[*table.AccountDetails, *table.AccountDetails]{Left: accountDetails[i], Right: nil})
			}
		}
		withHTMLData(c, "points", pairs)
	}
}

func refreshAccountDetailsHandler(engine *engine.Engine) gin.HandlerFunc {
	return checkAccountHandler(engine)
}

func collectAccountPointsHandler(engine *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		err = engine.CollectAccountPoints(id)
		if err != nil {
			withJsonError(c, http.StatusInternalServerError, err)
			return
		}
		withJSONData(c, true, "Collected account points successfully.")
	}
}
