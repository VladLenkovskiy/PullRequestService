package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pr-service/internal/handlers/dto"
	"sort"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	testDuration = 10 * time.Second
	baseURL      = "http://localhost:8080"
)

type requestType string

const (
	reqCreateTeam requestType = "create_team"
	reqGetTeam    requestType = "get_team"
	reqCreatePR   requestType = "create_pr"
)

type config struct {
	typ        requestType
	rps        int
	maxTimeout time.Duration
}

type result struct {
	typ     requestType
	latency time.Duration
	err     error
	status  int
}

const (
	load_tb = `DELETE FROM teams;
				INSERT INTO teams (name)
				SELECT 'team_' || g
				FROM generate_series(1, 100) AS g;

				DELETE FROM users;
				INSERT INTO users (user_id, username, team_name, is_active)
				SELECT
					'user_' || t || '_' || u AS user_id,
					'username_' || t || '_' || u AS username,
					'team_' || t AS team_name,
					TRUE AS is_active
				FROM generate_series(1, 100) AS t
				CROSS JOIN generate_series(1, 100) AS u;

				DELETE FROM pull_requests;
				INSERT INTO pull_requests (pr_id, pr_name, author_id, status)
				SELECT
					'pr_' || g AS pr_id,
					'PR number ' || g AS pr_name,
					'user_' || ((g-1)%100 + 1) || '_' || ((g-1)%100 + 1) AS author_id,
					'OPEN' AS status
				FROM generate_series(1, 10000) AS g;`
)

func main() {
	_ = godotenv.Load()

	mode := getEnv("LOAD_MODE", "load")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "pr_service")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	log.Println("DB connected")

	switch mode {
	case "load":
		if _, err := db.ExecContext(ctx, load_tb); err != nil {
			log.Fatalf("exec seed sql: %v", err)
		}

		log.Println("Seeding finished")
	case "test":
		testLoad()
		log.Println("load finished")
	case "clear":
		if err := clearTestData(db); err != nil {
			log.Fatalf("failed to clear test data: %v", err)
		}
		log.Println("Test data cleared successfully")
	default:
		log.Fatalf("unknown LOAD_MODE: %s", mode)
	}
}

func testLoad() {
	cfgs := []config{
		{typ: reqCreateTeam, rps: 5, maxTimeout: 300 * time.Millisecond},
		{typ: reqGetTeam, rps: 5, maxTimeout: 300 * time.Millisecond},
		{typ: reqCreatePR, rps: 5, maxTimeout: 300 * time.Millisecond},
	}

	log.Printf("Starting load test: duration=%s\n", testDuration)

	results := runParallelLoad(cfgs)

	summarizeAll(results, cfgs, testDuration)
}

func runParallelLoad(cfgs []config) []result {
	var wg sync.WaitGroup
	resultsCh := make(chan result, 10000)

	for _, cfg := range cfgs {
		cfg := cfg
		wg.Add(1)
		go func() {
			defer wg.Done()
			runLoadForConfig(cfg, resultsCh)
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var all []result
	for r := range resultsCh {
		all = append(all, r)
	}
	return all
}

func runLoadForConfig(cfg config, out chan<- result) {
	client := &http.Client{
		Timeout: cfg.maxTimeout,
	}

	interval := time.Second / time.Duration(cfg.rps)
	deadline := time.Now().Add(testDuration)

	log.Printf("[%-11s] RPS=%d, timeout=%s\n", cfg.typ, cfg.rps, cfg.maxTimeout)

	n := 0
	for time.Now().Before(deadline) {
		n++

		go func(iter int) {
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), cfg.maxTimeout)
			defer cancel()

			status, err := sendRequest(ctx, client, cfg.typ, iter)
			lat := time.Since(start)

			out <- result{
				typ:     cfg.typ,
				latency: lat,
				err:     err,
				status:  status,
			}
		}(n)

		time.Sleep(interval)
	}
}

func sendRequest(ctx context.Context, client *http.Client, typ requestType, n int) (int, error) {
	switch typ {
	case reqCreateTeam:
		return sendCreateTeam(ctx, client, n)
	case reqGetTeam:
		return sendGetTeam(ctx, client, n)
	case reqCreatePR:
		return sendCreatePR(ctx, client, n)
	default:
		return 0, fmt.Errorf("unknown request type: %s", typ)
	}
}

func sendGetTeam(ctx context.Context, client *http.Client, n int) (int, error) {
	teamIndex := (n % 50) + 1
	teamName := fmt.Sprintf("team_%d", teamIndex)

	url := fmt.Sprintf("%s/team/get?team_name=%s", baseURL, teamName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
}

func sendCreateTeam(ctx context.Context, client *http.Client, n int) (int, error) {
	url := baseURL + "/team/add"

	teamName := fmt.Sprintf("team_%d", n)
	bodyStruct := dto.CreateTeamIn{
		Name: teamName,
		Members: []dto.UserDTO{
			{ID: fmt.Sprintf("user_%d_1", n), Username: fmt.Sprintf("username_%d_1", n), IsActive: true},
			{ID: fmt.Sprintf("user_%d_2", n), Username: fmt.Sprintf("username_%d_2", n), IsActive: true},
		},
	}

	b, err := json.Marshal(bodyStruct)
	if err != nil {
		return 0, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func sendCreatePR(ctx context.Context, client *http.Client, n int) (int, error) {
	url := baseURL + "/pullRequest/create"

	authorID := "user_1_1"

	bodyStruct := dto.CreatePullRequestIn{
		ID:       fmt.Sprintf("load_pr_%d", n),
		Name:     fmt.Sprintf("Load PR %d", n),
		AuthorID: authorID,
	}

	b, err := json.Marshal(bodyStruct)
	if err != nil {
		return 0, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func summarizeAll(results []result, cfgs []config, duration time.Duration) {
	if len(results) == 0 {
		log.Println("No results collected")
		return
	}

	byType := map[requestType][]result{}
	for _, r := range results {
		byType[r.typ] = append(byType[r.typ], r)
	}

	fmt.Println("============================================")
	fmt.Println("              LOAD TEST SUMMARY             ")
	fmt.Println("============================================")
	fmt.Printf("Duration: %s\n", duration)
	fmt.Printf("Total requests: %d\n\n", len(results))

	types := make([]requestType, 0, len(byType))
	for t := range byType {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })

	for _, t := range types {
		summarizePerType(t, byType[t], duration)
		fmt.Println("--------------------------------------------")
	}

	var total, success, failed int
	var sumLatency time.Duration
	var maxLatency time.Duration

	for _, r := range results {
		total++
		if r.err != nil {
			failed++
		} else {
			success++
		}
		sumLatency += r.latency
		if r.latency > maxLatency {
			maxLatency = r.latency
		}
	}

	var avgLatency time.Duration
	if total > 0 {
		avgLatency = time.Duration(float64(sumLatency) / float64(total))
	}
	realRPS := float64(total) / duration.Seconds()

	fmt.Println("============ OVERALL =======================")
	fmt.Printf("Total:   %d\n", total)
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed:  %d\n", failed)
	fmt.Printf("Real RPS:        %.2f\n", realRPS)
	fmt.Printf("Avg latency:     %s\n", avgLatency)
	fmt.Printf("Max latency:     %s\n", maxLatency)
	fmt.Println("============================================")

}

func summarizePerType(typ requestType, results []result, duration time.Duration) {
	var (
		total, success, failed int
		sumLatency             time.Duration
		maxLatency             time.Duration
	)

	errSamples := map[string]int{}
	statusCounters := map[int]int{}

	for _, r := range results {
		total++
		statusCounters[r.status]++

		if r.err != nil {
			failed++
			if len(errSamples) < 5 {
				errSamples[r.err.Error()]++
			}
		} else {
			success++
		}

		sumLatency += r.latency
		if r.latency > maxLatency {
			maxLatency = r.latency
		}
	}

	var avgLatency time.Duration
	if total > 0 {
		avgLatency = time.Duration(float64(sumLatency) / float64(total))
	}
	realRPS := float64(total) / duration.Seconds()

	fmt.Printf("Type: %s\n", typ)
	fmt.Printf("  Requests:  %d\n", total)
	fmt.Printf("  Success:   %d\n", success)
	fmt.Printf("  Failed:    %d\n", failed)
	fmt.Printf("  Real RPS:  %.2f\n", realRPS)
	fmt.Printf("  Avg lat:   %s\n", avgLatency)
	fmt.Printf("  Max lat:   %s\n", maxLatency)

	if len(statusCounters) > 0 {
		fmt.Printf("  Status codes:\n")
		statuses := make([]int, 0, len(statusCounters))
		for s := range statusCounters {
			statuses = append(statuses, s)
		}
		sort.Ints(statuses)
		for _, s := range statuses {
			fmt.Printf("    %d: %d\n", s, statusCounters[s])
		}
	}

	if len(errSamples) > 0 {
		fmt.Println("  Sample errors:")
		for msg, count := range errSamples {
			fmt.Printf("    [%d] %s\n", count, truncate(msg, 120))
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func clearTestData(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tables := []string{
		"pull_request_reviewers",
		"pull_requests",
		"users",
		"teams",
	}

	for _, tbl := range tables {
		if _, err := tx.Exec("DELETE FROM " + tbl); err != nil {
			return fmt.Errorf("delete from %s: %w", tbl, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
