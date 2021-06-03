// +build broken

package virtual

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/entity"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/jeremyhahn/cropdroid/webservice/rest"
)

type VirtualRoom struct {
	ctx         common.Context
	port        int
	router      *mux.Router
	service     service.RoomService
	metricTable *common.MetricTable
	mapper      mapper.RoomMapper
	eventType   string
	closeChan   chan bool
}

func NewVirtualRoom(ctx common.Context, port int, roomService service.RoomService) *VirtualRoom {

	return &VirtualRoom{
		ctx:         ctx,
		port:        port,
		router:      mux.NewRouter().StrictSlash(true),
		service:     roomService,
		metricTable: common.NewMetricTable(common.CONTROLLER_TYPE_ROOM),
		mapper:      mapper.NewRoomMapper(),
		eventType:   "VirtualRoom",
		closeChan:   make(chan bool, 1)}
}

func (room *VirtualRoom) Run() {

	room.ctx.GetLogger().Debugf("[VirtualRoom] Starting virtual room on port %d", room.port)

	room.router.HandleFunc("/room", room.status).Methods("GET")
	room.router.HandleFunc("/reload", room.reload).Methods("GET")
	room.router.HandleFunc("/set/{metric}/{value}", room.setMetric).Methods("GET")

	room.loadState()

	http.ListenAndServe("localhost:8001", room.router)
}

func (room *VirtualRoom) status(w http.ResponseWriter, r *http.Request) {
	room.ctx.GetLogger().Debug("[VirtualRoom] /status")
	roomEntity := room.mapper.MapMetricTableToState(room.metricTable)
	rest.NewJsonWriter().Write(w, 200, roomEntity)
}

func (room *VirtualRoom) reload(w http.ResponseWriter, r *http.Request) {
	room.ctx.GetLogger().Debug("[VirtualRoom] /reload")
	room.loadState()
	roomEntity := room.mapper.MapMetricTableToState(room.metricTable)
	rest.NewJsonWriter().Write(w, 200, roomEntity)
}

func (room *VirtualRoom) setMetric(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	metric := params["metric"]
	value := params["value"]

	room.ctx.GetLogger().Debug("[VirtualRoom] /set/%s/%s", metric, value)

	valueFloat, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	room.metricTable.Put(metric, valueFloat)

	roomEntity := room.mapper.MapMetricTableToState(room.metricTable)
	rest.NewJsonWriter().Write(w, 200, roomEntity)
}

func (room *VirtualRoom) loadState() {
	fakeDataFile := "public_html/vroom.json"
	if _, err := os.Stat(fakeDataFile); err == nil {
		var entity entity.Room
		data, err := ioutil.ReadFile(fakeDataFile)
		if err != nil {
			room.ctx.GetLogger().Error(err.Error())
			return
		}
		err = json.Unmarshal(data, &entity)
		if err != nil {
			room.ctx.GetLogger().Error(err.Error())
			return
		}
		room.metricTable = room.mapper.MapStateToMetricTable(&entity)
	}
}
