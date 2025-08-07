package dbrepo

import (
	"errors"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"time"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("cannot insert a reservation")
	}

	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	if rr.RoomID == 3 {
		return errors.New("cannot insert a reservation")
	}

	return nil
}

func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(roomID int, start, end time.Time) (bool, error) {
	return false, nil
}

func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room

	if id > 2 {
		return room, errors.New("id out of range")
	}

	return room, nil
}

func (m *testDBRepo) GetUserById(userId int) (models.User, error) {
	var user models.User

	return user, nil
}

func (m *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, password string) (int, string, error) {
	return 1, "", nil
}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation
	return res, nil
}

func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (m *testDBRepo) UpdateProcessedForReservation(id int, processed int) error {
	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil
}

func (m *testDBRepo) GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction

	return restrictions, nil
}

func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

func (m *testDBRepo) DeleteBlockById(id int) error {
	return nil
}
