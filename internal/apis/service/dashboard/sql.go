package dashboard

import (
	"fmt"

	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/utils"
)

type prePareListSQLParams struct {
	Page      uint
	PageSize  uint
	Order     string
	UserID    uint
	TimeRange string
}

func (s *Service) prepareListSQL(params prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
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

	if params.Order == "favorites" {
		if params.TimeRange == "all" {
			sql := `
			SELECT
				tb1.*,
				tb2.username,
				tb2.username,
				tb2.email,
				tb2.icon_url,
				tb3.stared,
				tb3.user_id as stared_user_id,
				COUNT( tb3.dashboard_id ) AS favorites_count
			FROM
				%s AS tb1
				LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
				LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id
			WHERE
				tb1.is_privacy = FALSE
			GROUP BY
				tb1.id,
				tb2.username,
				tb2.email,
				tb2.icon_url,
				tb3.stared,
				stared_user_id	
			ORDER BY
				favorites_count DESC 
				LIMIT ? OFFSET ( ? - 1 ) * ?
			`
			sql = fmt.Sprintf(sql, tb1, tb2, tb3)
			queryRaw = s.db.Raw(sql, params.PageSize, params.Page, params.PageSize)

			countSql := `
			SELECT COUNT(tb1.id)
			FROM
				%s AS tb1
			WHERE
				tb1.is_privacy = FALSE 
			`
			countSql = fmt.Sprintf(countSql, tb1)
			countRaw = s.db.Raw(countSql)
			return
		}
		sql := `
		SELECT
			tb1.*,
			tb2.username,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			tb3.user_id as stared_user_id,
			COUNT( tb3.dashboard_id ) AS favorites_count
		FROM
			%s AS tb1
			LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id
		WHERE
			tb1.is_privacy = FALSE 
			AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP
		GROUP BY
			tb1.id,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			stared_user_id		
		ORDER BY
			favorites_count DESC 
			LIMIT ? OFFSET ( ? - 1 ) * ?
		`
		sql = fmt.Sprintf(sql, tb1, tb2, tb3, timeRangeFormat)
		queryRaw = s.db.Raw(sql, params.PageSize, params.Page, params.PageSize)

		countSql := `
		SELECT COUNT(tb1.id)
		FROM
			%s AS tb1
		WHERE
			tb1.is_privacy = FALSE 
			AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP
		`
		countSql = fmt.Sprintf(countSql, tb1, timeRangeFormat)
		countRaw = s.db.Raw(countSql)
		return
	}

	if params.Order == "new" {
		sql := `
		SELECT
			tb1.*,
			tb2.username,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			tb3.user_id as stared_user_id,
			COUNT( tb3.dashboard_id ) AS favorites_count
		FROM
			%s AS tb1
			LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id
		WHERE
			tb1.is_privacy = FALSE
		GROUP BY
			tb1.id,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			stared_user_id
		ORDER BY
			tb1.created_at DESC 
			LIMIT ? OFFSET ( ? - 1 ) * ?
		`
		sql = fmt.Sprintf(sql, tb1, tb2, tb3)
		queryRaw = s.db.Raw(sql, params.PageSize, params.Page, params.PageSize)

		countSql := `
		SELECT COUNT(tb1.id)
		FROM
			%s AS tb1
		WHERE
			tb1.is_privacy = FALSE 
		`
		countSql = fmt.Sprintf(countSql, tb1)
		countRaw = s.db.Raw(countSql)
		return
	}

	// TODO: default is trending
	if params.TimeRange == "all" {
		sql := `
		SELECT
			tb1.*,
			tb2.username,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			tb3.user_id as stared_user_id,
			COUNT( tb3.dashboard_id ) AS favorites_count
		FROM
			%s AS tb1
			LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id AND tb3.stared = TRUE
		WHERE
			tb1.is_privacy = FALSE
		GROUP BY
			tb1.id,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			stared_user_id
		ORDER BY
			tb1.updated_at DESC 
			LIMIT ? OFFSET ( ? - 1 ) * ?
		`
		sql = fmt.Sprintf(sql, tb1, tb2, tb3)
		queryRaw = s.db.Raw(sql, params.PageSize, params.Page, params.PageSize)

		countSql := `
		SELECT COUNT(tb1.id)
		FROM
			%s AS tb1
		WHERE
			tb1.is_privacy = FALSE 
		`
		countSql = fmt.Sprintf(countSql, tb1)
		countRaw = s.db.Raw(countSql)

	} else {
		sql := `
		SELECT
			tb1.*,
			tb2.username,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			tb3.user_id as stared_user_id,
			COUNT( tb3.dashboard_id ) AS favorites_count
		FROM
			%s AS tb1
			LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id AND tb3.stared = TRUE
		WHERE
			tb1.is_privacy = FALSE
			AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP
		GROUP BY
			tb1.id,
			tb2.username,
			tb2.email,
			tb2.icon_url,
			tb3.stared,
			stared_user_id
		ORDER BY
			tb1.updated_at DESC 
			LIMIT ? OFFSET ( ? - 1 ) * ?
		`
		sql = fmt.Sprintf(sql, tb1, tb2, tb3, timeRangeFormat)
		queryRaw = s.db.Raw(sql, params.PageSize, params.Page, params.PageSize)

		countSql := `
		SELECT COUNT(tb1.id)
		FROM
			%s AS tb1
		WHERE
			tb1.is_privacy = FALSE 
			AND tb1.created_at BETWEEN '%s' AND LOCALTIMESTAMP
		`
		countSql = fmt.Sprintf(countSql, tb1, timeRangeFormat)
		countRaw = s.db.Raw(countSql)
	}

	return
}

func (s *Service) prepateUserListSQL(params prePareListSQLParams) (queryRaw *gorm.DB, countRaw *gorm.DB, err error) {
	tb1 := datamodel.DashboardModel{}.TableName()
	tb2 := datamodel.UserModel{}.TableName()
	tb3 := datamodel.UserDashboardFavorites{}.TableName()
	sql := `
	SELECT
		tb1.*,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared
		tb3.user_id as stared_user_id
		COUNT( tb3.dashboard_id ) AS favorites_count
	FROM
		%s AS tb1
		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
		LEFT JOIN %s AS tb3 ON tb1.id = tb3.dashboard_id
	WHERE
		tb1.is_privacy = FALSE 
		AND tb1.user_id = ?
	GROUP BY
		tb1.id,
		tb2.username,
		tb2.email,
		tb2.icon_url,
		tb3.stared,
		stared_user_id
	ORDER BY
		favorites_count DESC 
		LIMIT ? OFFSET ( ? - 1 ) * ?
	`
	sql = fmt.Sprintf(sql, tb1, tb2, tb3)
	queryRaw = s.db.Raw(sql, params.UserID, params.PageSize, params.Page, params.PageSize)

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

func (s *Service) preparePopularDashboardTags(limit uint) (raw *gorm.DB, err error) {
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
