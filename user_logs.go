package hmrapi

import ( 
	"strings"
	//"strconv"
	"fmt"
	"github.com/osu-datenshi/api/common"
	"time"
	"gopkg.in/thehowl/go-osuapi.v1"
	"github.com/osu-datenshi/lib/getrank"
)

type Score struct {
	ID         int                  `json:"id"`
	Score      int64                `json:"score"`
	Mods       int                  `json:"mods"`
	Count300   int                  `json:"count_300"`
	Count100   int                  `json:"count_100"`
	Count50    int                  `json:"count_50"`
	CountMiss  int                  `json:"count_miss"`
	PlayMode   int                  `json:"play_mode"`
	Accuracy   float64              `json:"accuracy"`
	Rank       string               `json:"rank"`
}

type LogSimple struct {
	SongName	string	`json:"song_name"`
	LogBody		string	`json:"body"`
	Time		common.UnixTimestamp		`json:"time"`
	ScoreID		int		`json:"scoreid"`
	BeatmapID	int		`json:"beatmap_id"`
	Rank		string	`json:"rank"`
}

type Massive struct {
	common.ResponseBase
	Log		[]LogSimple	`json:"logs"`
}

func LogsGET(md common.MethodData) common.CodeMessager {
	id := md.Query("userid")
	mode := md.Query("mode")
	spmode := common.Int(md.Query("spmode"))

	results, err := md.DB.Query(fmt.Sprintf(`SELECT 
b.song_name, 
l.log, l.time, l.scoreid, 
b.beatmap_id, 
s.play_mode, s.mods, s.accuracy, s.300_count, s.100_count, s.50_count, s.misses_count 
FROM users_logs as l 
LEFT JOIN beatmaps as b USING (beatmap_md5) 
INNER JOIN scores_master as s ON s.id = l.scoreid 
WHERE user = %s 
AND l.game_mode = %s 
AND l.time > %d 
AND s.special_mode = %d 
ORDER BY l.time 
DESC LIMIT 10
`, id, mode, int(time.Now().Unix())-2592000, spmode))
	if err != nil {
		md.Err(err)
		return common.SimpleResponse(500, "Uh oh... Seems like Makino did something bad to API... Please try again! If it's broken... Please tell me in the Discord!")
	}

	var response Massive
	var logs []LogSimple

	defer results.Close()
	for results.Next() {
		var ls LogSimple
		var s Score
		results.Scan(
			&ls.SongName, &ls.LogBody, &ls.Time, &ls.ScoreID, &ls.BeatmapID,
			&s.PlayMode, &s.Mods, &s.Accuracy, &s.Count300, &s.Count100, &s.Count50, &s.CountMiss,
		)
		
		ls.Rank = strings.ToUpper(getrank.GetRank(
			osuapi.Mode(s.PlayMode),
			osuapi.Mods(s.Mods),
			s.Accuracy,
			s.Count300,
			s.Count100,
			s.Count50,
			s.CountMiss,
		))

		logs = append(logs, ls)
	}
	if err := results.Err(); err != nil {
		md.Err(err)
	}
	response.Log = logs
	response.Code = 200
	return response
}

//Thank you Kurikku!!
//https://github.com/osukurikku/api/blob/master/vendor/github.com/KotRikD/krapi/user_logs.go
