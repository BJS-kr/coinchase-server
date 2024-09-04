package test

import (
	"coin_chase/game"
	"coin_chase/http_server"
	"coin_chase/protodef"
	"coin_chase/worker_pool"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

// 프로그램의 random배치 특성상 유저의 정확한 순위 등을 예측하기는 어렵다.
// 프로젝트의 목적 자체에 가까운 high throughput, small packet size, synchronization등을 테스트한다.

// 유일하게 명시적으로 확인이 가능한 부분은 클라이언트 송신(send_client.go)이 100회 이상 실패시 terminate되는 부분이다.
// 이 부분 또한 추상화 되어 상술한 유효성 검사 하위에서 돌아가므로, 기능이 정상적으로 동작함을 확인했다.
// 또 한 각 자원(워커 및 그에 관련된 하위 goroutine들)의 생성 및 종료가 일원화 되어있는데,
// 워커의 반환은 테스트하지만 워커의 하위 자원들의 회수까지 테스트하지 않는다.
// 이러한 이유로, 이 테스트는 단순히 동시성 처리가 가능한 서버의 동작 여부를 확인하는 것에 가깝다.
// 이러한 테스트의 단점을 커버하기 위해, 실제로 클라이언트 프로그램(데스트탑 앱)을 작성해 테스트를 진행했다.
type TestClient struct {
	ID         string
	Conn       *net.UDPConn
	WorkerPort int
}

// 아래의 테스트는 순차적으로 진행되어야 하므로 실패하면 즉시 종료하기 위해 Fatalf로 처리합니다.
var (
	clientListeners []*TestClient
	testServer      *httptest.Server
)

func TestMain(m *testing.M) {
	clientListeners = make([]*TestClient, 0)
	testServer = httptest.NewServer(http_server.NewServer())

	defer testServer.Close()

	os.Exit(m.Run())
}

func TestInitialResources(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/server-state")

	if err != nil {
		t.Fatalf("failed to get server state: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	serverState := make(map[string]interface{})
	// json으로 받은 response를 map으로 변환

	json.NewDecoder(resp.Body).Decode(&serverState)

	workerCount := int(serverState["workerCount"].(float64))
	coinCount := int(serverState["coinCount"].(float64))
	itemCount := int(serverState["itemCount"].(float64))
	t.Run("워커 생성(프로그램이 켜질 때 함께 생성됨. 워커 갯수 검사)", func(t *testing.T) {
		if workerCount != worker_pool.WORKER_COUNT {
			t.Fatalf("worker pool initialization failed. expected: %d, go: %d", worker_pool.WORKER_COUNT, workerCount)
		}
	})
	// 기본 자원(map, coin, item) 생성 및 자원 갯수 검사
	t.Run("기본 자원 생성(맵, 코인, 아이템)", func(t *testing.T) {
		gameMap := game.GetGameMap()
		// 코인 검사(자원이 맵에 뿌려졌다는 것 자체가 맵이 잘 생성되었다는 것)
		if coinCount > game.COIN_COUNT || coinCount == 0 { // <= 조건인 이유는 코인은 랜덤성을 위하여 중복된 위치가 선정될 경우 그냥 스킵해버리기 때문에 COIN_COUNT보다 적게 생성될 수도 있다.
			t.Fatalf("coin count is not correct. expected: %d, got: %d", game.COIN_COUNT, len(gameMap.Coins))
		}

		// 아이템 검사
		if itemCount != game.ITEM_COUNT { // 코인과 다르게 아이템은 무조건 ITEM_COUNT만큼 생성되어야 한다.
			t.Fatalf("item count is not correct. expected: %d, got: %d", game.ITEM_COUNT, len(gameMap.RandomItems))
		}
	})
}

func TestWorkerPullOut(t *testing.T) {
	// 최대 수의 유저 로그인(워커풀이 비었음을 검사하고, 추가로 로그인 시도 시 실패)
	t.Run("최대 수의 유저(워커의 갯수 만큼) 로그인", func(t *testing.T) {
		// 최대 수의 유저 로그인
		for i := 0; i < worker_pool.WORKER_COUNT; i++ {
			// 로그인 시도
			// client는 UDP로 데이터를 전달 받기 때문에 먼저 UDP connection을 생성해야 한다.
			conn, err := net.ListenUDP("udp", &net.UDPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 0, // OS에게 빈 포트 요청
			})

			if err != nil {
				t.Fatalf("failed to create UDP connection: %s", err)
			}

			clientPort := conn.LocalAddr().(*net.UDPAddr).Port
			// 로그인 성공 검사
			// 유저 아이디 생성
			userId := "user" + strconv.Itoa(i)
			resp, err := http.Get(fmt.Sprintf(testServer.URL+"/get-worker-port/%s/%d", userId, clientPort))

			if err != nil || resp.StatusCode != http.StatusOK {
				t.Fatalf("failed to get worker port: %s", err)
			}

			defer resp.Body.Close()
			respBytes, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatalf("failed to read response body: %s", err)
			}

			workerPort, err := strconv.Atoi(string(respBytes))

			if err != nil {
				t.Fatalf("failed to convert worker port to int: %s", err)
			}
			// 로그인 성공 시 worker port를 받아온다.
			clientListeners = append(clientListeners, &TestClient{
				Conn:       conn,
				WorkerPort: workerPort,
				ID:         userId,
			})
		}
	})

	t.Run("초과 로그인 시 로그인 실패", func(t *testing.T) {
		userId := "user-over-limit"
		resp, _ := http.Get(fmt.Sprintf(testServer.URL+"/get-worker-port/%s/%d", userId, 0))

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected status code: %d, got: %d", http.StatusConflict, resp.StatusCode)
		}
	})
}

func TestPlay(t *testing.T) {
	// 유저의 랜덤 위치 이동(스코어가 쌓일 시간을 주기 위해 10초 실행. 복수의 유저가 크래시 없이 game map과 scoreboard에 write하고 점수를 쌓는 것 자체가 테스트의 목적)
	// 정확한 데이터는 예측할 수 없다. 랜덤으로 동전들이 움직일 뿐 더러, 유저의 이동도 랜덤이기 때문에 모두가 0점일 수도 있고 점수가 같거나 높을 수도 있다.
	t.Run("유저의 랜덤 위치 이동", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		var allClientWaitGroup sync.WaitGroup
		timeoutReached := false
		overallTimeLimit := 10

		go func() {
			time.Sleep(time.Second * time.Duration(overallTimeLimit))
			timeoutReached = true
		}()

		for _, testClient := range clientListeners {
			allClientWaitGroup.Add(1)
			go func() {
				userPosition := &protodef.Position{
					X: 0,
					Y: 0,
				}

				for !timeoutReached {
					// ! 중요 !
					// 원래대로라면 서버가 클라이언트의 위치를 강제로 결정할 수 있다(클라이언트가 유효하지 않은 위치에 존재할 시).
					// 그래서 서버가 보내주는 데이터를 토대로 클라이언트에서 유저의 위치를 재조정하는 등의 동작이 가능하지만럼(멀티 플레이 게임에서 렉 걸릴 때 캐릭터 위치 강제로 결정되는 것 처럼)
					// 이를 구현하려면 클라이언트에 구현되어 있는 것 처럼,
					// 각 테스트 클라이언트가 서버의 UDP 통신을 수신하고 다시 그에 따라 위치를 position을 결정하는
					// 클라이언트의 로직이 서버 테스트에 포함되게 되므로, 서버의 테스트가 아닌 클라이언트의 테스트가 되어버린다.
					// 서버가 애초에 out of bound 혹은 이미 점거된 위치에 대한 판단을 서버가 알아서 진행하므로, 유효하지 않은 위치를 계산하지 않고
					// 클라이언트의 랜덤 포지션을 그대로 전송한다. 이 방법으로도 데이터에 대한 경합성 해결, 요청 무효화 등의 테스트는 가능하다.
					newUserPosition := &protodef.Position{
						X: userPosition.X + game.GenerateRandomDirection(),
						Y: userPosition.Y + game.GenerateRandomDirection(),
					}

					userStatus := &protodef.Status{
						Id:              testClient.ID,
						CurrentPosition: newUserPosition,
					}

					userPosition = newUserPosition
					marshaledData, err := proto.Marshal(userStatus)

					if err != nil {
						panic(err)
					}
					// 유저의 위치를 서버에 전송
					_, err = testClient.Conn.WriteToUDP(marshaledData, &net.UDPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: testClient.WorkerPort,
					})

					if err != nil {
						t.Log(err)
					}

					time.Sleep(time.Millisecond * 200)
				}

				allClientWaitGroup.Done()
			}()

			allClientWaitGroup.Wait()
		}
	})
}

func TestWorkerPutBack(t *testing.T) {
	t.Run("로그아웃(워커 반환)", func(t *testing.T) {
		for _, testClient := range clientListeners {
			req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf(testServer.URL+"/disconnect/%s", testClient.ID), nil)

			if err != nil {
				t.Fatalf("failed to create request: %s", err)
			}

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				t.Fatalf("failed to put worker port back: %s", err.Error())
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
			}

			defer resp.Body.Close()
		}
	})
}
