package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients sync.Map
var broadcast = make(chan PixelUpdate)

// PixelUpdate structure for broadcasting pixel changes
type PixelUpdate struct {
    X     int    `json:"x"`
    Y     int    `json:"y"`
    Color string `json:"color"`
}

func main() {
    // Initialize the database
    initDB()

    // Set up routes
    http.Handle("/", http.FileServer(http.Dir("./static")))
    http.HandleFunc("/ws", handleConnections)
    http.HandleFunc("/save-pixel", savePixelHandler)
    http.HandleFunc("/clear-data", clearDataHandler)

    // Start the message handling goroutine
    go handleMessages()

    // Start the server
    fmt.Println("Server started on :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

// handleConnections upgrades HTTP connections to WebSocket and manages new clients
func handleConnections(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Error upgrading connection: %v", err)
        return
    }
    defer conn.Close()

    clients.Store(conn, true)

    // Listen for incoming messages from the client
    for {
        var update PixelUpdate
        err := conn.ReadJSON(&update)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            clients.Delete(conn)
            break
        }

        // Save the pixel update to the database
        err = savePixel(update.X, update.Y, update.Color)
        if err != nil {
            log.Printf("Database save error: %v", err)
            continue
        }

        // Broadcast the update to all clients
        broadcast <- update
    }
}

// handleMessages broadcasts any incoming update to all connected clients
func handleMessages() {
    for {
        update := <-broadcast
        clients.Range(func(key, value interface{}) bool {
            client := key.(*websocket.Conn)
            err := client.WriteJSON(update)
            if err != nil {
                log.Printf("Error writing message: %v", err)
                client.Close()
                clients.Delete(client)
            }
            return true
        })
    }
}

// savePixelHandler handles HTTP POST requests to save a pixel
func savePixelHandler(w http.ResponseWriter, r *http.Request) {
    var req PixelUpdate
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := savePixel(req.X, req.Y, req.Color)
    if err != nil {
        http.Error(w, "Failed to save pixel", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(req)
}

// clearDataHandler clears all pixel data from the database
func clearDataHandler(w http.ResponseWriter, r *http.Request) {
    clearDatabase()
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("All data cleared"))
}

