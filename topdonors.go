package hmrapi

import (
	"github.com/osu-datenshi/api/common"
)

type userData struct {
	ID             int                  `json:"id"`
	Username       string               `json:"username"`
	UsernameAKA    string               `json:"username_aka"`
	RegisteredOn   common.UnixTimestamp `json:"registered_on"`
	LatestActivity common.UnixTimestamp `json:"latest_activity"`
	Country        string               `json:"country"`
	Expiration     common.UnixTimestamp `json:"expiration"`
}

type topDonorsResponse struct {
	common.ResponseBase
	Users []userData `json:"users"`
}

const lbUserQuery = `
SELECT
	users.id, users.username, master_stats.username_aka, users.register_datetime, users.privileges, users.latest_activity,
	master_stats.country, users.donor_expire
FROM users
INNER JOIN master_stats ON master_stats.id = users.id
WHERE users.privileges >= 4 AND users.privileges != 1048576
ORDER BY users.donor_expire DESC
`

func TopDonorsGET(md common.MethodData) common.CodeMessager {
	var resp topDonorsResponse
	resp.Code = 200

	var tempUsers []userData

	rows, err := md.DB.Query(lbUserQuery)
	if err != nil {
		md.Err(err)
		return common.SimpleResponse(500, "Uh oh... Seems like Aoba did something bad to API... Please try again! If it's broken... Please tell me in the Discord!")
	}
	defer rows.Close()
	for rows.Next() {
		var u userData
		var privileges uint64
		err := rows.Scan(
			&u.ID, &u.Username, &u.UsernameAKA, &u.RegisteredOn, &privileges, &u.LatestActivity,
			&u.Country, &u.Expiration,
		)
		if err != nil {
			md.Err(err)
			continue
		}

		var HasDonor, IsCheat bool
		HasDonor = common.UserPrivileges(privileges)&common.UserPrivilegeDonor > 0
		IsCheat = common.UserPrivileges(privileges)&common.AdminPrivilegeAccessRAP > 0
		if IsCheat {
			continue
		}
		if HasDonor {
			tempUsers = append(tempUsers, u)
		} else {
			continue
		}
	}

	if len(tempUsers)>100 {
		sortedUsers := make([]userData, 100)
		copy(sortedUsers, tempUsers)
		resp.Users = sortedUsers
	} else {
		resp.Users = tempUsers
	}
	return resp
}

//Thank you Kurikku!!
//https://github.com/osukurikku/api/blob/master/vendor/github.com/KotRikD/krapi/topdonors.go
