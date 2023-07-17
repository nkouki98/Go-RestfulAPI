package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

//Aircraft Struct M(Model)

type Aircraft struct {
	ID              string    `json:"id"`
	Manufacturer    string    `json:"manufacturer"`
	Model           string    `json:"model"`
	LastMaintenance time.Time `json:"lastMaintenance"`
	Age             int       `json:"age"`
	Leased          bool      `json:"leased"`
	Status          string    `json:"status"`
	SeatCapacity    int       `json:"seatcapacity"`
	LastFlown       time.Time `json:"lastflown"`
}

// slice or collection of aircracts
var aircrafts []Aircraft

func getAircrafts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT id, manufacturer, model, lastMaintenance, age, leased, status, seatcapacity, lastflown FROM aircraft")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var aircrafts []Aircraft
	for rows.Next() {
		var aircraft Aircraft
		err := rows.Scan(&aircraft.ID, &aircraft.Manufacturer, &aircraft.Model, &aircraft.LastMaintenance, &aircraft.Age, &aircraft.Leased, &aircraft.Status, &aircraft.SeatCapacity, &aircraft.LastFlown)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		aircrafts = append(aircrafts, aircraft)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(aircrafts)
}

func getAircraft(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	aircraftID := params["id"]
	row := db.QueryRow("SELECT id, manufacturer, model, lastMaintenance, age, leased, status, seatcapacity, lastflown FROM aircraft WHERE id = ?", aircraftID)
	var aircraft Aircraft

	err := row.Scan(&aircraft.ID, &aircraft.Manufacturer, &aircraft.Model, &aircraft.LastMaintenance, &aircraft.Age, &aircraft.Leased, &aircraft.Status, &aircraft.SeatCapacity, &aircraft.LastFlown)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(aircraft)
}

func createAircraft(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var aircraft Aircraft
	_ = json.NewDecoder(r.Body).Decode(&aircraft)
	insertQuery := "INSERT INTO aircraft (id, manufacturer, model, lastMaintenance, age, leased, status, seatcapacity, lastflown) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	_, err := db.Exec(insertQuery, aircraft.ID, aircraft.Manufacturer, aircraft.Model, aircraft.LastMaintenance, aircraft.Age, aircraft.Leased, aircraft.Status, aircraft.SeatCapacity, aircraft.LastFlown)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(aircraft)
}

func deleteAircraft(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	aircraftID := params["id"]

	// DB delete
	deleteQuery := "DELETE FROM aircraft WHERE id = ?"
	_, err := db.Exec(deleteQuery, aircraftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	index := -1
	for i, aircraft := range aircrafts {
		if aircraft.ID == aircraftID {
			index = i
			break
		}
	}
	if index == -1 {
		http.NotFound(w, r)
		return
	}

	// Remove the aircraft from the slice
	aircrafts = append(aircrafts[:index], aircrafts[index+1:]...)

	json.NewEncoder(w).Encode(aircrafts)
}

func updateAircraft(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	aircraftID := params["id"]
	updateQuery := "UPDATE aircraft SET manufacturer=?, model=?, lastMaintenance=?, age=?, leased=?, status=?, seatcapacity=?, lastflown=? WHERE id=?"
	var updatedAircraft *Aircraft
	for i := range aircrafts {
		if aircrafts[i].ID == aircraftID {
			updatedAircraft = &aircrafts[i]
			break
		}
	}
	if updatedAircraft == nil {
		http.NotFound(w, r)
		return
	}
	err := json.NewDecoder(r.Body).Decode(updatedAircraft)
	if err != nil {
		http.Error(w, "Invalid data provided", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(updateQuery, updatedAircraft.Manufacturer, updatedAircraft.Model, updatedAircraft.LastMaintenance, updatedAircraft.Age, updatedAircraft.Leased, updatedAircraft.Status, updatedAircraft.SeatCapacity, updatedAircraft.LastFlown, aircraftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedAircraft)
}

var db *sql.DB

func main() {

	var err error
	// Open the database connection and assign it to the 'db' variable
	db, err = sql.Open("mysql", "root:admin@tcp(127.0.0.1:3306)/aircraft_management?parseTime=true")
	if err != nil {
		log.Fatalf("impossible to create the connection: %s", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping the database: %s", err)
	}
	// Initialize mux router
	r := mux.NewRouter()

	// dummy data for testing without database
	aircrafts = append(aircrafts, Aircraft{
		ID:              "AC001",
		Manufacturer:    "Boeing",
		Model:           "Boeing 777",
		LastMaintenance: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
		Age:             5,
		Leased:          true,
		Status:          "active",
		LastFlown:       time.Date(2023, time.July, 10, 8, 0, 0, 0, time.UTC),
		SeatCapacity:    345,
	})

	aircrafts = append(aircrafts, Aircraft{
		ID:              "AC005",
		Manufacturer:    "Airbus",
		Model:           "Airbus A320 NEO",
		LastMaintenance: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
		Age:             2,
		Leased:          false,
		Status:          "active",
		LastFlown:       time.Date(2023, time.July, 10, 8, 0, 0, 0, time.UTC),
		SeatCapacity:    189,
	})

	r.HandleFunc("/api/aircrafts", getAircrafts).Methods("GET")           // GET all aircrafts
	r.HandleFunc("/api/aircrafts/{id}", getAircraft).Methods("GET")       // GET an aircraft by ID
	r.HandleFunc("/api/aircrafts", createAircraft).Methods("POST")        // Create a new aircraft
	r.HandleFunc("/api/aircrafts/{id}", deleteAircraft).Methods("DELETE") // Delete an aircraft by ID
	r.HandleFunc("/api/aircrafts/{id}", updateAircraft).Methods("PUT")    // Update an aircraft by ID

	log.Fatal(http.ListenAndServe(":8000", r))
}
