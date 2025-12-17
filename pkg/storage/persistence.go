package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// SaveToFile сохраняет все сессии в файл
func (s *SessionStore) SaveToFile() error {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()

	if !s.persistEnabled {
		return nil
	}

	data := s.preparePersistData()

	if err := s.saveDataToFile(data, s.persistFile); err != nil {
		return err
	}

	s.lastSaveTime = time.Now()
	s.dirty = false

	s.logger.Debug("Sessions saved to file",
		zap.String("file", s.persistFile),
		zap.Int("session_count", data.TotalCount),
		zap.Time("saved_at", data.SavedAt),
		zap.Duration("session_max_age", data.SessionMaxAge),
	)

	return nil
}

// LoadFromFile загружает сессии из файла
func (s *SessionStore) LoadFromFile() error {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()

	if !s.persistEnabled {
		return nil
	}

	file, err := os.Open(s.persistFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func(file *os.File) {
		errClose := file.Close()
		if errClose != nil {
			s.logger.Error("Error closing file", zap.Error(errClose))
		}
	}(file)

	var data PersistedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// Используем сохраненный sessionMaxAge или значение по умолчанию
	var maxAge time.Duration
	if data.SessionMaxAge > 0 {
		maxAge = data.SessionMaxAge
	} else {
		maxAge = defaultSessionMaxAge
	}

	// Фильтруем просроченные сессии при загрузке
	now := time.Now()
	validSessions := make(map[string]*SessionInfo)
	for id, session := range data.Sessions {
		// Не загружаем отозванные сессии
		if session.IsRevoked {
			continue
		}

		// Проверяем возраст сессии с учетом сохраненного максимального возраста
		if now.Sub(session.LastUsedAt) > maxAge {
			continue
		}

		validSessions[id] = session
	}

	s.mu.Lock()
	s.sessions = validSessions
	s.sessionMaxAge = maxAge
	s.mu.Unlock()

	s.dirty = false
	s.lastSaveTime = time.Now()

	s.logger.Info("Sessions loaded from file",
		zap.String("file", s.persistFile),
		zap.Int("loaded_count", len(validSessions)),
		zap.Int("original_count", data.TotalCount),
		zap.Duration("session_max_age", maxAge),
	)

	return nil
}

// preparePersistData подготавливает данные для сохранения
func (s *SessionStore) preparePersistData() *PersistedData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := &PersistedData{
		Version:       storeVersion,
		SavedAt:       time.Now(),
		SessionMaxAge: s.sessionMaxAge,
		Sessions:      make(map[string]*SessionInfo, len(s.sessions)),
		TotalCount:    len(s.sessions),
	}

	// Копируем сессии
	for k, v := range s.sessions {
		data.Sessions[k] = &SessionInfo{
			UserID:     v.UserID,
			CreatedAt:  v.CreatedAt,
			LastUsedAt: v.LastUsedAt,
			IsRevoked:  v.IsRevoked,
			UserAgent:  v.UserAgent,
			IPAddress:  v.IPAddress,
		}
	}

	return data
}

// saveDataToFile сохраняет данные в файл
func (s *SessionStore) saveDataToFile(data *PersistedData, filename string) error {
	// Создаем директорию если ее нет
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// Создаем временный файл для атомарной записи
	tempFile := filename + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		errClose := file.Close()
		if errClose != nil {
			s.logger.Error("Failed to close temp file", zap.Error(errClose))
		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if errEncode := encoder.Encode(data); errEncode != nil {
		errRemove := os.Remove(tempFile)
		if errRemove != nil {
			return errRemove
		}
		return errEncode
	}

	// Закрываем файл перед переименованием
	errClose := file.Close()
	if errClose != nil {
		return errClose
	}

	// Атомарно заменяем старый файл новым
	if errRename := os.Rename(tempFile, filename); errRename != nil {
		errRemove := os.Remove(tempFile)
		if errRemove != nil {
			return errRemove
		}
		return errRename
	}

	return nil
}

// Backup создает резервную копию сессий
func (s *SessionStore) Backup(backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = "backups"
	}

	// Создаем директорию для бэкапов
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	// Генерируем имя файла с временной меткой
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(backupDir, "sessions_backup_"+timestamp+".json")

	// Подготавливаем данные для сохранения
	data := s.preparePersistData()

	// Сохраняем данные в файл бэкапа
	if err := s.saveDataToFile(data, backupFile); err != nil {
		return "", err
	}

	s.logger.Info("Backup created",
		zap.String("backup_file", backupFile),
		zap.Int("session_count", data.TotalCount),
		zap.Duration("session_max_age", data.SessionMaxAge),
	)

	return backupFile, nil
}

// EnablePersistence включает/выключает сохранение на диск
func (s *SessionStore) EnablePersistence(enabled bool) {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()

	s.persistEnabled = enabled
	if enabled {
		s.saveTicker.Reset(defaultSaveInterval)
	} else {
		s.saveTicker.Stop()
	}

	s.logger.Info("Persistence enabled", zap.Bool("enabled", enabled))
}

// SetPersistFile устанавливает файл для сохранения
func (s *SessionStore) SetPersistFile(filename string) error {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()

	if filename == "" || filename == s.persistFile {
		return nil
	}

	// Сохраняем текущие данные в новый файл
	oldFile := s.persistFile
	s.persistFile = filename

	if err := s.SaveToFile(); err != nil {
		s.persistFile = oldFile // Восстанавливаем старый файл при ошибке
		return err
	}

	s.logger.Info("Persist file changed",
		zap.String("old_file", oldFile),
		zap.String("new_file", filename),
	)

	return nil
}

// ForceSave принудительно сохраняет сессии на диск
func (s *SessionStore) ForceSave() error {
	return s.SaveToFile()
}

// GetPersistInfo возвращает информацию о состоянии сохранения
func (s *SessionStore) GetPersistInfo() map[string]interface{} {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()

	info := make(map[string]interface{})
	info["persist_enabled"] = s.persistEnabled
	info["persist_file"] = s.persistFile
	info["last_save_time"] = s.lastSaveTime
	info["is_dirty"] = s.dirty
	info["save_interval"] = defaultSaveInterval.String()

	return info
}

// MarkDirty помечает хранилище как измененное
func (s *SessionStore) MarkDirty() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}
