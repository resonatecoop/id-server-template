package oauth

import (
	"errors"

	"github.com/RichardKnop/go-oauth2-server/models"
)

// FindUserMetaValue finds user meta value by meta_key and user_id
func (s *Service) FindWpUserMetaValue(userId uint64, key string) (string, error) {
	wpusermeta := new(models.WpUserMeta)
	notFound := s.db2.Table("rsntr_usermeta").Select("meta_value").Where("user_id = ? AND meta_key = ?", userId, key).
		First(wpusermeta).RecordNotFound()

	// Not found
	if notFound {
		return "", errors.New("Data not found")
	}

	return wpusermeta.MetaValue, nil
}

// UpdateUserMetaValue update or create wp user meta value
func (s *Service) UpdateWpUserMetaValue(userId uint64, key string, value string) error {
	wpusermeta := &models.WpUserMeta{
		MetaKey:   key,
		MetaValue: value,
		UserId:    userId,
	}

	if s.db2.Model(&wpusermeta).
		Where("user_id = ? AND meta_key = ?", userId, key).
		Updates(&wpusermeta).RowsAffected == 0 {
		return s.db2.Create(wpusermeta).Error
	}
	return nil
}
