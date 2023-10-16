package models

import (
	"database/sql"
	"errors"
	"time"
)

type JobModelInterface interface {
	Insert(userID int, company, role, commute, applicationStatus, location, notes string, dateApplied time.Time) (int, error)
	Delete(jobID int) error
	Update(jobID int, company, role, commute, applicationStatus, location, notes string, dateApplied time.Time) error
	Get(jobID int) (*Job, error)
	GetLatest(userID int) ([]*Job, error)
	GetAll(userID int) ([]*Job, error)
	UserMatchesJob(userID int, jobID int) (bool, error)
}

type Job struct {
	ID                int
	UserID            int
	Company           string
	Role              string
	Commute           string
	ApplicationStatus string
	Location          string
	DateApplied       time.Time
	Notes             string
}

type JobModel struct {
	DB *sql.DB
}

func (m *JobModel) Insert(userID int, company, role, commute, applicationStatus, location, notes string, dateApplied time.Time) (int, error) {
	stmt := `INSERT INTO jobs (user_id, company, job_role, commute, application_status, location, notes, date_applied) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, userID, company, role, commute, applicationStatus, location, notes, dateApplied)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *JobModel) Delete(jobID int) error {
	stmt := `DELETE FROM jobs WHERE id = ?`

	_, err := m.DB.Exec(stmt, jobID)

	return err
}

func (m *JobModel) Update(jobID int, company, role, commute, applicationStatus, location, notes string, dateApplied time.Time) error {
	stmt := `UPDATE jobs SET company = ?, job_role = ?, commute = ?, application_status = ?, location = ?, notes = ?, date_applied = ?
	WHERE id = ?`

	_, err := m.DB.Exec(stmt, company, role, commute, applicationStatus, location, notes, dateApplied, jobID)

	return err
}

func (m *JobModel) Get(jobID int) (*Job, error) {
	stmt := `SELECT user_id, company, job_role, commute, application_status, location, date_applied, notes 
	FROM jobs WHERE id = ?`

	j := &Job{}

	err := m.DB.QueryRow(stmt, jobID).Scan(&j.UserID, &j.Company, &j.Role, &j.Commute, &j.ApplicationStatus, &j.Location, &j.DateApplied, &j.Notes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	j.ID = jobID

	return j, nil
}

func (m *JobModel) GetLatest(userID int) ([]*Job, error) {
	stmt := `SELECT id, company, job_role, commute, application_status, location, date_applied, notes 
	FROM jobs WHERE user_id = ? ORDER by id DESC LIMIT 15`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	jobs := []*Job{}

	for rows.Next() {
		j := &Job{}

		err = rows.Scan(&j.ID, &j.Company, &j.Role, &j.Commute, &j.ApplicationStatus, &j.Location, &j.DateApplied, &j.Notes)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (m *JobModel) GetAll(userID int) ([]*Job, error) {
	stmt := `SELECT id, company, job_role, commute, application_status, location, date_applied, notes 
	FROM jobs WHERE user_id = ? ORDER BY date_applied DESC`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	jobs := []*Job{}

	for rows.Next() {
		j := &Job{}

		err = rows.Scan(&j.ID, &j.Company, &j.Role, &j.Commute, &j.ApplicationStatus, &j.Location, &j.DateApplied, &j.Notes)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (m *JobModel) UserMatchesJob(userID int, jobID int) (bool, error) {
	var uid int

	stmt := `SELECT user_id FROM jobs WHERE id = ?`

	err := m.DB.QueryRow(stmt, jobID).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNoRecord
		} else {
			return false, err
		}
	}

	return uid == userID, err
}
