package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1rowvy/routy-example-gin/internal/model"
)

// store keeps everything in memory: restart the service to start over.
type store struct {
	mu       sync.Mutex
	users    map[int]*model.User
	orders   map[int]*model.Order
	tokens   map[string]int
	sessions map[string]int
	idem     map[string]int
	nextID   int
}

var catalog = map[string]float64{"BOOK-1": 12.5, "MUG-2": 8, "TEE-3": 20}

const adminEmail = "admin@shop.test"

func newStore() *store {
	s := &store{
		users:    map[int]*model.User{},
		orders:   map[int]*model.Order{},
		tokens:   map[string]int{},
		sessions: map[string]int{},
		idem:     map[string]int{},
		nextID:   1,
	}
	s.users[1] = &model.User{ID: 1, Name: "Admin", Email: adminEmail, Role: "admin", CreatedAt: time.Now().UTC()}
	s.nextID = 2
	return s
}

// adminPassword is ADMIN_PASSWORD, "admin" by default.
func adminPassword() string {
	if p := os.Getenv("ADMIN_PASSWORD"); p != "" {
		return p
	}
	return "admin"
}

func randomHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func currentUser(c *gin.Context) model.User {
	return c.MustGet("user").(model.User)
}

func (s *store) id() int {
	id := s.nextID
	s.nextID++
	return id
}

func (s *store) checkPassword(email, password string) (model.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Every user except the admin has the password "customer".
	for _, u := range s.users {
		if u.Email == email && ((u.Role == "admin" && password == adminPassword()) || (u.Role != "admin" && password == "customer")) {
			return *u, true
		}
	}
	return model.User{}, false
}

func (s *store) newToken(userID int) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := randomHex()
	s.tokens[t] = userID
	return t
}

func (s *store) newSession(userID int) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := randomHex()
	s.sessions[t] = userID
	return t
}

func (s *store) lookup(m map[string]int, key string) (model.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[m[key]]
	if !ok {
		return model.User{}, false
	}
	return *u, true
}

func (s *store) userByToken(t string) (model.User, bool)   { return s.lookup(s.tokens, t) }
func (s *store) userBySession(t string) (model.User, bool) { return s.lookup(s.sessions, t) }

func (s *store) createUser(req model.CreateUser) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Email == req.Email {
			return model.User{}, errors.New("email is taken")
		}
	}
	role := req.Role
	if role == "" {
		role = "customer"
	}
	u := &model.User{ID: s.id(), Name: req.Name, Email: req.Email, Role: role, CreatedAt: time.Now().UTC()}
	s.users[u.ID] = u
	return *u, nil
}

func (s *store) user(id string) (model.User, bool) {
	n, _ := strconv.Atoi(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[n]
	if !ok {
		return model.User{}, false
	}
	return *u, true
}

func (s *store) setAvatar(id, path string) (model.User, bool) {
	n, _ := strconv.Atoi(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[n]
	if !ok {
		return model.User{}, false
	}
	u.Avatar = &path
	return *u, true
}

func (s *store) deleteUser(id string) bool {
	n, _ := strconv.Atoi(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[n]; !ok || n == 1 {
		return false
	}
	delete(s.users, n)
	return true
}

func (s *store) createOrder(req model.CreateOrder, key string) (model.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.idem[key]; ok && key != "" {
		return *s.orders[id], nil
	}
	if _, ok := s.users[req.UserID]; !ok {
		return model.Order{}, errors.New("no such user")
	}
	if len(req.Items) == 0 {
		return model.Order{}, errors.New("an order needs at least one item")
	}
	o := &model.Order{ID: s.id(), UserID: req.UserID, Status: "new", Note: req.Note, Created: time.Now().UTC()}
	for _, it := range req.Items {
		price, ok := catalog[it.SKU]
		if !ok {
			return model.Order{}, errors.New("unknown sku " + it.SKU)
		}
		it.Price = price
		o.Items = append(o.Items, it)
		o.Total += price * float64(it.Qty)
	}
	s.orders[o.ID] = o
	if key != "" {
		s.idem[key] = o.ID
	}
	return *o, nil
}

func (s *store) order(id string) (model.Order, bool) {
	n, _ := strconv.Atoi(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[n]
	if !ok {
		return model.Order{}, false
	}
	return *o, true
}

func (s *store) listOrders(userID int, status string, page int) model.OrderPage {
	s.mu.Lock()
	defer s.mu.Unlock()
	var all []model.Order
	for _, o := range s.orders {
		if (o.UserID == userID || s.users[userID].Role == "admin") && (status == "" || o.Status == status) {
			all = append(all, *o)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID > all[j].ID })
	if page < 1 {
		page = 1
	}
	from, to := min((page-1)*20, len(all)), min(page*20, len(all))
	return model.OrderPage{Items: append([]model.Order{}, all[from:to]...), Page: page, Total: len(all)}
}

// pay marks the order pending; a payment provider "confirms" it 1.5s later.
func (s *store) pay(id string) (model.Order, error) {
	n, _ := strconv.Atoi(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[n]
	if !ok {
		return model.Order{}, errors.New("no such order")
	}
	if o.Status != "new" {
		return model.Order{}, errors.New("order is already " + o.Status)
	}
	o.Status = "pending"
	time.AfterFunc(1500*time.Millisecond, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		now := time.Now().UTC()
		o.Status, o.PaidAt = "paid", &now
	})
	return *o, nil
}
