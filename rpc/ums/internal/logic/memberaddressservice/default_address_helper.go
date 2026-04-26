package memberaddressservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"gorm.io/gorm"
)

func ensureMemberHasDefaultAddress(ctx context.Context, tx *query.Query, memberID int64) error {
	q := tx.UmsMemberAddress
	defaultCount, err := q.WithContext(ctx).
		Where(q.MemberID.Eq(memberID), q.IsDeleted.Eq(0), q.IsDefault.Eq(1)).
		Count()
	if err != nil {
		return err
	}
	if defaultCount > 0 {
		return nil
	}

	item, err := q.WithContext(ctx).
		Where(q.MemberID.Eq(memberID), q.IsDeleted.Eq(0)).
		Order(q.CreateTime.Desc(), q.ID.Desc()).
		First()
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return err
	}

	_, err = q.WithContext(ctx).Where(q.ID.Eq(item.ID)).Update(q.IsDefault, 1)
	return err
}

func clearMemberDefaultAddresses(ctx context.Context, tx *query.Query, memberID int64) error {
	q := tx.UmsMemberAddress
	_, err := q.WithContext(ctx).
		Where(q.MemberID.Eq(memberID), q.IsDeleted.Eq(0), q.IsDefault.Eq(1)).
		Update(q.IsDefault, 0)
	return err
}
