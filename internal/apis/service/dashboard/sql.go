package dashboard

import (
	"fmt"

	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/utils"
)

type prePareListSQLParams struct {
	Page          uint
	PageSize      uint
	Order         string
	UserID        uint
	CurrentUserId uint
	TimeRange     string
}

func (s *Service) prepareListSQL(params *prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
	timeRangeFormat := ""
	if params.TimeRange != "all" {
		if timeRangeFormat, err = utils.FormatTimeRange(params.TimeRange); err != nil {
			return nil, nil, err
		}

	}
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserModel{}.TableName()
	tb3 := datamodel.UserDashboardFavorites{}.TableName()

	if params.UserID != 0 {
		return s.prepateUserListSQL(params)
	}

	prepareSql := `
	SELECT
		tb1.*,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared,
		COUNT ( tb4.dashboard_id ) AS favorites_count 
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.ID 
		LEFT JOIN %s AS tb3 ON tb1.ID = tb3.dashboard_id 
			AND tb3.user_id = ?
		LEFT JOIN %s AS tb4 ON tb1.ID = tb4.dashboard_id
			AND tb4.stared = TRUE  
	WHERE
		tb1.is_privacy = FALSE
		%s
	GROUP BY
		tb1.ID,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared 
	ORDER BY
		%s
	LIMIT ? OFFSET ( ? - 1 ) * ?
	`

	prepareCountSql := `
	SELECT COUNT(tb1.id)
	FROM
		%s AS tb1
	WHERE
		tb1.is_privacy = FALSE
		%s 
	`

	if params.Order == "favorites" {
		if params.TimeRange == "all" {
			sql := fmt.Sprintf(prepareSql,
				tb1,
				tb2,
				tb3,
				tb3,                    // tb4
				"",                     // time range
				"favorites_count DESC", // order by
			)

			queryRaw = s.db.Raw(sql, params.CurrentUserId, params.PageSize, params.Page, params.PageSize)

			prepareCountSql = fmt.Sprintf(prepareCountSql, tb1, "")
			countRaw = s.db.Raw(prepareCountSql)
			return
		}

		sql := fmt.Sprintf(prepareSql,
			tb1,
			tb2,
			tb3,
			tb3, // tb4
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP",
				timeRangeFormat), // time range
			"favorites_count DESC ", // order by
		)
		queryRaw = s.db.Raw(sql, params.CurrentUserId, params.PageSize, params.Page, params.PageSize)

		fmt.Println(queryRaw.Statement.SQL.String())
		countSql := fmt.Sprintf(prepareCountSql, tb1,
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP",
				timeRangeFormat), // time range
		)
		countRaw = s.db.Raw(countSql)
		return
	}

	if params.Order == "new" {
		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3,                  // tb4
			"",                   // time range
			"tb1.created_at ASC", // order by
		)
		queryRaw = s.db.Raw(sql, params.CurrentUserId, params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, "")
		countRaw = s.db.Raw(countSql)
		return
	}

	// TODO: default is trending
	if params.TimeRange == "all" {
		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3,                   // tb4
			"",                    // time range
			"tb1.updated_at DESC", // order by
		)
		queryRaw = s.db.Raw(sql, params.CurrentUserId, params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, "")
		countRaw = s.db.Raw(countSql)

	} else {

		sql := fmt.Sprintf(prepareSql,
			tb1,
			tb2,
			tb3,
			tb3, // tb4
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP",
				timeRangeFormat), // time range
			"tb1.updated_at DESC ", // order by
		)
		queryRaw = s.db.Raw(sql, params.CurrentUserId, params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1,
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP",
				timeRangeFormat), // time range
		)
		countRaw = s.db.Raw(countSql)
	}

	return
}

func (s *Service) prepateUserListSQL(params *prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserModel{}.TableName()
	tb3 := datamodel.UserDashboardFavorites{}.TableName()
	sql := `
	SELECT
		tb1.*,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared,
		COUNT( tb4.dashboard_id ) AS favorites_count
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
		LEFT JOIN %s AS tb3 ON tb1.ID = tb3.dashboard_id 
			AND tb3.user_id = ?
		LEFT JOIN %s AS tb4 ON tb1.ID = tb4.dashboard_id
			AND tb4.stared = TRUE  
	WHERE
		tb1.is_privacy = FALSE 
		AND tb1.user_id = ?
	GROUP BY
		tb1.id,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared
	ORDER BY
		favorites_count DESC 
		LIMIT ? OFFSET ( ? - 1 ) * ?
	`
	sql = fmt.Sprintf(sql, tb1, tb2, tb3, tb3)
	queryRaw = s.db.Raw(sql, params.UserID, params.UserID, params.PageSize, params.Page, params.PageSize)

	countSql := `
	SELECT COUNT(tb1.id)
	FROM
		%s AS tb1
	WHERE
		tb1.is_privacy = FALSE 
		AND tb1.user_id = ?
	`
	countSql = fmt.Sprintf(countSql, tb1)
	countRaw = s.db.Raw(countSql, params.UserID)
	return
}

func (s *Service) prepateBrowseUserListSQL(params *prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserModel{}.TableName()
	tb3 := datamodel.UserDashboardFavorites{}.TableName()
	sql := `
	SELECT
		tb1.*,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared,
		COUNT( tb4.dashboard_id ) AS favorites_count
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
		LEFT JOIN %s AS tb3 ON tb1.ID = tb3.dashboard_id 
			AND tb3.user_id = ?
		LEFT JOIN %s AS tb4 ON tb1.ID = tb4.dashboard_id
			AND tb4.stared = TRUE  
	WHERE
		tb1.is_privacy = FALSE 
		AND tb1.user_id = ?
	GROUP BY
		tb1.id,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared
	ORDER BY
		favorites_count DESC 
		LIMIT ? OFFSET ( ? - 1 ) * ?
	`
	sql = fmt.Sprintf(sql, tb1, tb2, tb3, tb3)
	queryRaw = s.db.Raw(sql,
		params.CurrentUserId, // guest user for stared
		params.UserID,        // access user for filter dashboard
		params.PageSize,
		params.Page,
		params.PageSize,
	)

	countSql := `
	SELECT COUNT(tb1.id)
	FROM
		%s AS tb1
	WHERE
		tb1.is_privacy = FALSE 
		AND tb1.user_id = ?
	`
	countSql = fmt.Sprintf(countSql, tb1)
	countRaw = s.db.Raw(countSql, params.UserID)
	return
}

func (s *Service) preparePopularDashboardTagsSQL(limit uint) (raw *gorm.DB, err error) {
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserDashboardFavorites{}.TableName()
	sql := `
	SELECT
		COUNT( tb2.dashboard_id ) AS favorites_count,
		tb1.tags as tags
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb1.id = tb2.dashboard_id
	WHERE
		tb1.is_privacy = FALSE 
	GROUP BY
		tb1.id
	ORDER BY
		favorites_count DESC
	LIMIT ?
	`
	sql = fmt.Sprintf(sql, tb1, tb2)
	raw = s.db.Raw(sql, limit)
	return raw, nil

}

func (s *Service) prepareListStaredSQL(params *prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserModel{}.TableName()
	tb3 := datamodel.UserDashboardFavorites{}.TableName()

	timeRangeFormat := ""
	if params.TimeRange != "all" {
		if timeRangeFormat, err = utils.FormatTimeRange(params.TimeRange); err != nil {
			return nil, nil, err
		}

	}

	prepareSql := `
	SELECT
		tb1.*,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared,
		COUNT( tb4.dashboard_id ) AS favorites_count
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb2.id = tb1.user_id
		LEFT JOIN %s AS tb3 ON tb3.dashboard_id = tb1.id 
			AND tb3.user_id = ?
			AND tb3.stared = TRUE
		LEFT JOIN %s AS tb4 ON tb1.ID = tb4.dashboard_id
			AND tb4.stared = TRUE
	WHERE
		tb1.is_privacy = FALSE
		AND tb3.stared = TRUE
		%s 
	GROUP BY
		tb1.id,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared
	ORDER BY
		%s
		LIMIT ? OFFSET ( ? - 1 ) * ?
	
	`

	prepareCountSql := `
	SELECT COUNT(tb1.id)
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb3 ON tb3.user_id = ? 
			AND tb1.id = tb3.dashboard_id 
			AND tb3.stared = TRUE
	WHERE
		tb1.is_privacy = FALSE
		AND tb3.stared = TRUE
		%s
	`

	if params.Order == "favorites" {
		if params.TimeRange == "all" {
			sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
				tb3,                     // tb4 for count stars
				"",                      // time range empty
				"favorites_count DESC ", // order by
			)
			queryRaw = s.db.Raw(sql,
				params.CurrentUserId, // tb3.user_id
				params.PageSize, params.Page, params.PageSize)

			countSql := fmt.Sprintf(prepareCountSql, tb1, tb3,
				"", // time range
			)
			countRaw = s.db.Raw(countSql,
				params.CurrentUserId, // tb3.user_id
			)
			return
		}

		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3, // tb4 for count stars
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP", timeRangeFormat), // time range
			"favorites_count DESC ", // order by
		)
		queryRaw = s.db.Raw(sql,
			params.CurrentUserId, // tb3.user_id
			params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, tb3,
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP", timeRangeFormat)) // time range
		countRaw = s.db.Raw(countSql,
			params.CurrentUserId, // tb3.user_id
		)
		return
	}

	if params.Order == "new" {
		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3,                    // tb4 for count stars
			"",                     // time range
			"tb1.created_at DESC ", // order by
		)
		queryRaw = s.db.Raw(sql,
			params.CurrentUserId, // tb3.user_id
			params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, tb3,
			"", // time range
		)
		countRaw = s.db.Raw(countSql,
			params.CurrentUserId, // tb3.user_id
		)
		return
	}

	// TODO: default is trending
	if params.TimeRange == "all" {
		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3,                    // tb4 for count stars
			"",                     // time range
			"tb1.updated_at DESC ", // order by
		)
		queryRaw = s.db.Raw(sql,
			params.CurrentUserId, // tb3.user_id
			params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, tb3,
			"", // time range
		)
		countRaw = s.db.Raw(countSql,
			params.CurrentUserId, // tb3.user_id
		)

	} else {
		sql := fmt.Sprintf(prepareSql, tb1, tb2, tb3,
			tb3, // tb4 for count stars
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP", timeRangeFormat), // time range
			"tb1.updated_at DESC ", // order by
		)
		queryRaw = s.db.Raw(sql,
			params.CurrentUserId, // tb3.user_id
			params.PageSize, params.Page, params.PageSize)

		countSql := fmt.Sprintf(prepareCountSql, tb1, tb3,
			fmt.Sprintf("AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP", timeRangeFormat)) // time range

		countRaw = s.db.Raw(countSql,
			params.CurrentUserId, // tb3.user_id
		)
	}

	return
}
