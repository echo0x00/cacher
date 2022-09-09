package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type Cache interface {
	Get() chan int
	Set(value int)
}

type cache struct {
	val        int
	exp        time.Time
	inProgress bool
	res        chan int
}

func (c *cache) Set(value int) {
	c.val = value
	c.exp = time.Now().Add(time.Second * 10)
}

func (c *cache) Get() chan int {
	exp := c.exp

	if time.Now().Unix() > exp.Unix() {
		if c.inProgress {
			return c.res
		}
		go func() {
			c.inProgress = true
			//slow func imitation
			time.Sleep(200 * time.Millisecond)
			r := rand.Intn(220)
			c.Set(r)
			c.inProgress = false
			c.res <- r
		}()
		return c.res
	} else {
		go func() {
			c.res <- c.val
		}()
		return c.res
	}

}

func NewCache() *cache {
	return &cache{
		val: 0,
		res: make(chan int),
	}
}

var o = fmt.Print

func main() {
	c := NewCache()

	http.HandleFunc("/pogoda/", func(w http.ResponseWriter, r *http.Request) {
		//pId := r.URL.Query().Get("pid")
		ctx, cf := context.WithTimeout(r.Context(), time.Millisecond*500)
		pogoda := getPogoda(c)
		defer cf()
		select {
		case pg := <-pogoda:
			io.WriteString(w, fmt.Sprintf("cache: %d", pg))
		case <-ctx.Done():
			io.WriteString(w, "Проверяем погоду, зайдите позже")
			break
		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "home")
	})

	err := http.ListenAndServe(":80", nil)

	if err != nil {
		panic(err.Error())
	}
}

func getPogoda(cache Cache) chan int {
	return cache.Get()
}
