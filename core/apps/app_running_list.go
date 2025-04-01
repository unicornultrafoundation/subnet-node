package apps

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-datastore"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

const RUNNING_APP_LIST_KEY = "running-apps"

func (s *Service) GetRunningAppListProto(ctx context.Context) (*pbapp.AppRunningList, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getRunningAppListProto(ctx)
}

func (s *Service) getRunningAppListProto(ctx context.Context) (*pbapp.AppRunningList, error) {
	runningAppListKey := datastore.NewKey(RUNNING_APP_LIST_KEY)
	runningAppListData, err := s.Datastore.Get(ctx, runningAppListKey)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			// Return an empty list instead of an error
			return &pbapp.AppRunningList{AppIds: [][]byte{}}, nil
		}
		return nil, err
	}

	var runningAppList pbapp.AppRunningList
	if err := proto.Unmarshal(runningAppListData, &runningAppList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal running app list: %w", err)
	}

	return &runningAppList, nil
}

func (s *Service) SaveRunningAppListProto(ctx context.Context, appList *pbapp.AppRunningList) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveRunningAppListProto(ctx, appList)
}

func (s *Service) saveRunningAppListProto(ctx context.Context, appList *pbapp.AppRunningList) error {
	runningAppList, err := proto.Marshal(appList)
	if err != nil {
		return fmt.Errorf("failed to marshal running app list: %w", err)
	}

	runningAppListKey := datastore.NewKey(RUNNING_APP_LIST_KEY)
	if err := s.Datastore.Put(ctx, runningAppListKey, runningAppList); err != nil {
		return fmt.Errorf("failed to save running app list: %w", err)
	}

	return nil
}

func (s *Service) AddNewRunningApp(ctx context.Context, appId *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	runningAppList, err := s.getRunningAppListProto(ctx)
	if err != nil {
		return fmt.Errorf("failed to add new running app: %v", appId)
	}

	appIdBytes := appId.Bytes()

	if isRunningAppExist(runningAppList, appIdBytes) {
		return nil
	}

	// Append the new appId to the list
	runningAppList.AppIds = append(runningAppList.AppIds, appIdBytes)

	// Save the updated list
	err = s.saveRunningAppListProto(ctx, runningAppList)
	if err != nil {
		return fmt.Errorf("failed to update running app list: %w", err)
	}

	return nil
}

func (s *Service) RemoveRunningApp(ctx context.Context, appId *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	runningAppList, err := s.getRunningAppListProto(ctx)
	if err != nil {
		return fmt.Errorf("failed to add new running app: %v", appId)
	}

	appIdBytes := appId.Bytes()

	if !isRunningAppExist(runningAppList, appIdBytes) {
		return nil
	}

	// Filter out the appId from the list
	newAppIds := make([][]byte, 0, len(runningAppList.AppIds))
	for _, id := range runningAppList.AppIds {
		if !bytes.Equal(id, appIdBytes) {
			newAppIds = append(newAppIds, id)
		}
	}

	// Update the list
	runningAppList.AppIds = newAppIds

	// Save the updated list
	err = s.saveRunningAppListProto(ctx, runningAppList)
	if err != nil {
		return fmt.Errorf("failed to update running app list: %w", err)
	}

	return nil
}

func (s *Service) ClearRunningAppListProto(ctx context.Context) error {
	emptyList := &pbapp.AppRunningList{AppIds: [][]byte{}} // Create an empty list

	// Save the empty list
	err := s.SaveRunningAppListProto(ctx, emptyList)
	if err != nil {
		return fmt.Errorf("failed to update running app list: %w", err)
	}

	return nil
}

func isRunningAppExist(appListData *pbapp.AppRunningList, appIdBytes []byte) bool {
	for _, id := range appListData.AppIds {
		if bytes.Equal(id, appIdBytes) {
			return true
		}
	}

	return false
}
