package todo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type List struct {
	Items []Item `json:"items"`
	Path  string `json:"-"`
}

func NewList(path string) (*List, error) {
	list := &List{
		Items: []Item{},
		Path:  path,
	}
	
	if err := list.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading todo list: %w", err)
		}
	}
	
	return list, nil
}

func (l *List) Add(title, description string) *Item {
	item := Item{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	l.Items = append(l.Items, item)
	return &item
}

func (l *List) Complete(id string) error {
	for i := range l.Items {
		if l.Items[i].ID == id {
			l.Items[i].Completed = true
			l.Items[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

func (l *List) Remove(id string) error {
	for i, item := range l.Items {
		if item.ID == id {
			l.Items = append(l.Items[:i], l.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

func (l *List) Load() error {
	if l.Path == "" {
		return fmt.Errorf("no path specified")
	}
	
	data, err := ioutil.ReadFile(l.Path)
	if err != nil {
		return err
	}
	
	if len(data) == 0 {
		l.Items = []Item{}
		return nil
	}
	
	return json.Unmarshal(data, l)
}

func (l *List) Save() error {
	if l.Path == "" {
		return fmt.Errorf("no path specified")
	}
	
	dir := filepath.Dir(l.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling todo list: %w", err)
	}
	
	return ioutil.WriteFile(l.Path, data, 0644)
}

func (l *List) String() string {
	if len(l.Items) == 0 {
		return "No todo items"
	}
	
	var result string
	for _, item := range l.Items {
		status := " "
		if item.Completed {
			status = "✓"
		}
		result += fmt.Sprintf("[%s] %s: %s\n", status, item.ID[:8], item.Title)
		if item.Description != "" {
			result += fmt.Sprintf("    %s\n", item.Description)
		}
	}
	return result
}