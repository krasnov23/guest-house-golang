package dbrepo

import (
	"context"
	"errors"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {

	// Автоматическая отмена запроса, если он выполняется дольше 3 секунд
	// Защита от "зависания" при проблемах с БД
	// Освобождение ресурсов (соединений с БД) при таймауте
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var newID int

	stmt := `insert into reservations (first_name, last_name,email,phone,start_date,
                          end_date,room_id,created_at,updated_at)
    		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	// Выполняется запрос с контекстом (для таймаута)
	// Передаются параметры из объекта res (бронирование) и текущее время для полей created/updated
	// Результат (новый ID) сканируется в переменную newID
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	stmt := `insert into room_restrictions (start_date,end_date, room_id,reservation_id,
				created_at,updated_at,restriction_id) VALUES ($1, $2, $3, $4,$5,$6,$7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		rr.StartDate,
		rr.EndDate,
		rr.RoomID,
		rr.ReservationID,
		time.Now(),
		time.Now(),
		rr.RestrictionID)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(roomID int, start, end time.Time) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var numRows int

	query := `select count(*) from reservations where 
                                      room_id = $1 and end_date > $2 and start_date < $3;`

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)

	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}

	// Возвращает true в случае если на отправленные даты комната свободна
	if numRows == 0 {
		return true, nil
	}

	// возвращает false если комната занята
	return false, nil
}

func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var rooms []models.Room

	query := `select r.id,r.room_name from rooms r where r.id not in (select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date);`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	// rows.Next() перебирает каждую строку результата.
	for rows.Next() {
		var room models.Room

		// rows.Scan копирует значения из строки в поля структуры room:
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)

		// Если ошибка (например, тип данных не совпал), возвращает собранные данные и ошибку.
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	// Когда срабатывает rows.Err()?
	// Проблемы с соединением - База данных упала во время итерации.
	// Некорректные данные - Например, повреждение сетевого пакета при передаче.
	// Если контекст отменился (таймаут или ручная отмена), но rows.Next() уже вернул false.
	// Строка if err = rows.Err(); err != nil обязательна для корректной обработки ошибок, которые:
	// Возникают между вызовами rows.Next().
	// Не связаны напрямую с парсингом данных (rows.Scan).
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var room models.Room

	query := `select r.id,r.room_name,r.created_at,r.updated_at from rooms r where r.id=$1;`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `select id,first_name,last_name,email, password, access_level,created_at,updated_at
			from users u where u.id=$1;`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `update users set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5 where id = $5;`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now(), u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	/*hashed, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Ошибка при хэшировании:", err)
	} else {
		fmt.Println("Хэш пароля:", string(hashed))
	}*/

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, `select id,password from users where email = $1`, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// проверка совпадает ли захэшированный пароль с базы с нашим отправленным паролем
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id,r.first_name,r.last_name,r.email,r.phone, r.start_date,
		r.end_date, r.room_id,r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Reservation

		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var res models.Reservation

	query := `
	select r.id,r.first_name,r.last_name,r.email,r.phone, r.start_date,
	r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from reservations r
	left join rooms rm on (r.room_id = rm.id)
	where r.id = $1;`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id,r.first_name,r.last_name,r.email,r.phone, r.start_date,
		r.end_date, r.room_id,r.created_at, r.updated_at,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.processed = 0
		order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Reservation

		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5 where id = $6;`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.Phone, time.Now(), u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `delete from reservations where id = $1;`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil

}

func (m *postgresDBRepo) UpdateProcessedForReservation(id int, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `update reservations set processed = $1 where id = $2;`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var rooms []models.Room

	query := `select id,room_name,created_at,updated_at from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Room
		err := rows.Scan(
			&i.ID,
			&i.RoomName,
			&i.CreatedAt,
			&i.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, i)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `select id,coalesce(reservation_id, 0),restriction_id,room_id,start_date,end_date
			  from room_restrictions where end_date > $1 and $2 >= start_date and room_id = $3`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.RoomRestriction
		err := rows.Scan(
			&i.ID,
			// поскольку reservationID ждет int а нам может прийти null, для этого используем функцию coalesce которая приведет к нулю пустое значение
			&i.ReservationID,
			&i.RestrictionID,
			&i.RoomID,
			&i.StartDate,
			&i.EndDate,
		)

		if err != nil {
			return nil, err
		}

		restrictions = append(restrictions, i)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}

func (m *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date,room_id,restriction_id, created_at, updated_at) values ($1, $2, $3, $4,$5,$6)`

	_, err := m.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())

	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func (m *postgresDBRepo) DeleteBlockById(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `delete from room_restrictions where id = $1;`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}
