package hmrapi

import (
	"fmt"
	"strings"

	"gopkg.in/thehowl/go-osuapi.v1"
	"github.com/osu-datenshi/api/common"
	"github.com/osu-datenshi/getrank"
)

type miniUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type Score2 struct {
	ID         int                  `json:"id"`
	BeatmapMD5 string               `json:"beatmap_md5"`
	Score      int64                `json:"score"`
	MaxCombo   int                  `json:"max_combo"`
	FullCombo  bool                 `json:"full_combo"`
	Mods       int                  `json:"mods"`
	Count300   int                  `json:"count_300"`
	Count100   int                  `json:"count_100"`
	Count50    int                  `json:"count_50"`
	CountGeki  int                  `json:"count_geki"`
	CountKatu  int                  `json:"count_katu"`
	CountMiss  int                  `json:"count_miss"`
	Time       common.UnixTimestamp `json:"time"`
	PlayMode   int                  `json:"play_mode"`
	Accuracy   float64              `json:"accuracy"`
	PP         float32              `json:"pp"`
	Rank       string               `json:"rank"`
	Completed  int                  `json:"completed"`
}

type MixedBeatmap struct {
	Score2
	Beatmap beatmap  `json:"beatmap"`
	User    miniUser `json:"user"`
}

type ScoresResponse struct {
	common.ResponseBase
	Scores []MixedBeatmap `json:"scores"`
}

const topPlaysQuery = `
SELECT
	s.id, s.beatmap_md5, s.score,
	s.max_combo, s.full_combo, s.mods,
	s.300_count, s.100_count, s.50_count,
	s.gekis_count, s.katus_count, s.misses_count,
	s.time, s.play_mode, s.accuracy, s.pp,
	s.completed,

	b.beatmap_id, b.beatmapset_id, b.beatmap_md5,
	b.song_name, b.ar, b.od, b.difficulty_std,
	b.difficulty_taiko, b.difficulty_ctb, b.difficulty_mania,
	b.max_combo, b.hit_length, b.ranked,
	b.ranked_status_freezed, b.latest_update,

	u.id, u.username
FROM scores_master as s
INNER JOIN beatmaps as b ON b.beatmap_md5 = s.beatmap_md5
INNER JOIN users as u ON u.id = s.userid
WHERE s.pp > 0 AND s.completed = '3' AND u.privileges & 1 > 0 AND s.play_mode = %s AND s.special_mode = %d
ORDER BY s.pp DESC
`

func TopPlaysGET(md common.MethodData) common.CodeMessager {
	limit := md.HasQuery("l")
	limitQuery := " LIMIT 50"
	if limit {
		limitQuery = " LIMIT " + md.Query("l")
	}
	mode := md.Query("mode")
	smode := common.Int(md.Query("spmode"))

	rows, err := md.DB.Query(fmt.Sprintf(topPlaysQuery, mode, smode) + limitQuery)
	if err != nil {
		md.Err(err)
		return common.SimpleResponse(500, "Uh oh... Seems like Makino did something bad to API... Please try again! If it's broken... Please tell me in the Discord!")
	}
	var scores []MixedBeatmap
	for rows.Next() {
		var (
			us MixedBeatmap
			b  beatmap
		)
		err = rows.Scan(
			&us.ID, &us.BeatmapMD5, &us.Score2.Score,
			&us.MaxCombo, &us.FullCombo, &us.Mods,
			&us.Count300, &us.Count100, &us.Count50,
			&us.CountGeki, &us.CountKatu, &us.CountMiss,
			&us.Time, &us.PlayMode, &us.Accuracy, &us.PP,
			&us.Completed,

			&b.BeatmapID, &b.BeatmapsetID, &b.BeatmapMD5,
			&b.SongName, &b.AR, &b.OD, &b.Diff2.STD,
			&b.Diff2.Taiko, &b.Diff2.CTB, &b.Diff2.Mania,
			&b.MaxCombo, &b.HitLength, &b.Ranked,
			&b.RankedStatusFrozen, &b.LatestUpdate,

			&us.User.ID, &us.User.Username,
		)
		if err != nil {
			md.Err(err)
			return common.SimpleResponse(500, "Uh oh... Seems like Makino did something bad to API... Please try again! If it's broken... Please tell me in the Discord!")
		}
		b.Difficulty = b.Diff2.STD
		us.Beatmap = b
		us.Rank = strings.ToUpper(getrank.GetRank(
			osuapi.Mode(us.PlayMode),
			osuapi.Mods(us.Mods),
			us.Accuracy,
			us.Count300,
			us.Count100,
			us.Count50,
			us.CountMiss,
		))
		scores = append(scores, us)
	}
	r := ScoresResponse{}
	r.Code = 200
	r.Scores = scores
	return r

}

//Thank you Kurikku!!
//https://github.com/osukurikku/api/blob/master/vendor/github.com/KotRikD/krapi/top_plays.go
