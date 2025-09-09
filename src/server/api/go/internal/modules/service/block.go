package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/memodb-io/Acontext/internal/modules/model"
	"github.com/memodb-io/Acontext/internal/modules/repo"
)

type BlockService interface {
	CreatePage(ctx context.Context, b *model.Block) error
	DeletePage(ctx context.Context, spaceID uuid.UUID, pageID uuid.UUID) error
	GetPageProperties(ctx context.Context, pageID uuid.UUID) (*model.Block, error)
	UpdatePageProperties(ctx context.Context, b *model.Block) error
	ListPageChildren(ctx context.Context, pageID uuid.UUID) ([]model.Block, error)
	MovePage(ctx context.Context, pageID uuid.UUID, newParentID *uuid.UUID, targetSort *int64) error
	UpdatePageSort(ctx context.Context, pageID uuid.UUID, sort int64) error

	CreateBlock(ctx context.Context, b *model.Block) error
	DeleteBlock(ctx context.Context, spaceID uuid.UUID, blockID uuid.UUID) error
	GetBlockProperties(ctx context.Context, blockID uuid.UUID) (*model.Block, error)
	UpdateBlockProperties(ctx context.Context, b *model.Block) error
	ListBlockChildren(ctx context.Context, blockID uuid.UUID) ([]model.Block, error)
	MoveBlock(ctx context.Context, blockID uuid.UUID, newParentID uuid.UUID, targetSort *int64) error
	UpdateBlockSort(ctx context.Context, blockID uuid.UUID, sort int64) error
}

type blockService struct{ r repo.BlockRepo }

func NewBlockService(r repo.BlockRepo) BlockService { return &blockService{r: r} }

func (s *blockService) CreatePage(ctx context.Context, b *model.Block) error {
	if b.Type == "" {
		b.Type = model.BlockTypePage
	}
	if b.Type != model.BlockTypePage {
		return errors.New("type must be page")
	}
	// Validate parent type: when parent_id is provided, it must be a page
	if b.ParentID != nil {
		parent, err := s.r.Get(ctx, *b.ParentID)
		if err != nil {
			return err
		}
		if parent.Type != model.BlockTypePage {
			return errors.New("parent must be page")
		}
	}
	next, err := s.r.NextSort(ctx, b.SpaceID, b.ParentID)
	if err != nil {
		return err
	}
	b.Sort = next
	return s.r.Create(ctx, b)
}

func (s *blockService) DeletePage(ctx context.Context, spaceID uuid.UUID, pageID uuid.UUID) error {
	if len(pageID) == 0 {
		return errors.New("page id is empty")
	}
	return s.r.Delete(ctx, spaceID, pageID)
}

func (s *blockService) GetPageProperties(ctx context.Context, pageID uuid.UUID) (*model.Block, error) {
	if len(pageID) == 0 {
		return nil, errors.New("page id is empty")
	}
	return s.r.Get(ctx, pageID)
}

func (s *blockService) UpdatePageProperties(ctx context.Context, b *model.Block) error {
	if len(b.ID) == 0 {
		return errors.New("page id is empty")
	}
	return s.r.Update(ctx, b)
}

func (s *blockService) ListPageChildren(ctx context.Context, pageID uuid.UUID) ([]model.Block, error) {
	if len(pageID) == 0 {
		return nil, errors.New("page id is empty")
	}
	return s.r.ListChildren(ctx, pageID)
}

func (s *blockService) MovePage(ctx context.Context, pageID uuid.UUID, newParentID *uuid.UUID, targetSort *int64) error {
	if len(pageID) == 0 {
		return errors.New("page id is empty")
	}
	// Validate parent type for moving: when newParentID is provided, it must be a page
	if newParentID != nil {
		parent, err := s.r.Get(ctx, *newParentID)
		if err != nil {
			return err
		}
		if parent.Type != model.BlockTypePage {
			return errors.New("new parent must be page")
		}
	}
	if targetSort == nil {
		return s.r.MoveToParentAppend(ctx, pageID, newParentID)
	}
	return s.r.MoveToParentAtSort(ctx, pageID, newParentID, *targetSort)
}

func (s *blockService) UpdatePageSort(ctx context.Context, pageID uuid.UUID, sort int64) error {
	if len(pageID) == 0 {
		return errors.New("page id is empty")
	}
	return s.r.ReorderWithinGroup(ctx, pageID, sort)
}

func (s *blockService) CreateBlock(ctx context.Context, b *model.Block) error {
	if b.Type == "" {
		b.Type = model.BlockTypeBlock
	}
	if b.Type != model.BlockTypeBlock {
		return errors.New("type must be block")
	}
	if b.ParentID == nil {
		return errors.New("parent id is required for block")
	}
	next, err := s.r.NextSort(ctx, b.SpaceID, b.ParentID)
	if err != nil {
		return err
	}
	b.Sort = next
	return s.r.Create(ctx, b)
}

func (s *blockService) DeleteBlock(ctx context.Context, spaceID uuid.UUID, blockID uuid.UUID) error {
	if len(blockID) == 0 {
		return errors.New("block id is empty")
	}
	return s.r.Delete(ctx, spaceID, blockID)
}

func (s *blockService) GetBlockProperties(ctx context.Context, blockID uuid.UUID) (*model.Block, error) {
	if len(blockID) == 0 {
		return nil, errors.New("block id is empty")
	}
	return s.r.Get(ctx, blockID)
}

func (s *blockService) UpdateBlockProperties(ctx context.Context, b *model.Block) error {
	if len(b.ID) == 0 {
		return errors.New("block id is empty")
	}
	return s.r.Update(ctx, b)
}

func (s *blockService) ListBlockChildren(ctx context.Context, blockID uuid.UUID) ([]model.Block, error) {
	if len(blockID) == 0 {
		return nil, errors.New("block id is empty")
	}
	return s.r.ListChildren(ctx, blockID)
}

func (s *blockService) MoveBlock(ctx context.Context, blockID uuid.UUID, newParentID uuid.UUID, targetSort *int64) error {
	if len(blockID) == 0 {
		return errors.New("block id is empty")
	}
	if targetSort == nil {
		return s.r.MoveToParentAppend(ctx, blockID, &newParentID)
	}
	return s.r.MoveToParentAtSort(ctx, blockID, &newParentID, *targetSort)
}

func (s *blockService) UpdateBlockSort(ctx context.Context, blockID uuid.UUID, sort int64) error {
	if len(blockID) == 0 {
		return errors.New("block id is empty")
	}
	return s.r.ReorderWithinGroup(ctx, blockID, sort)
}
