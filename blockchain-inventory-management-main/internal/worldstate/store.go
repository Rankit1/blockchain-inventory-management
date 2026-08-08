package worldstate

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

const WorldStateBucketName = "worldstate"

// Asset lifecycle states.
const (
	LifecycleActive      = "ACTIVE"
	LifecycleMaintenance = "MAINTENANCE"
	LifecycleIdle        = "IDLE"
	LifecycleRetired     = "RETIRED"
)

type Asset struct {
	AssetID          string    `json:"assetId"`
	DeptID           string    `json:"deptId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int       `json:"qty"`
	Threshold        int       `json:"threshold"`
	PriorityTier     string    `json:"priorityTier,omitempty"`
	CriticalityScore float64   `json:"criticalityScore,omitempty"`
	LifecycleState   string    `json:"lifecycleState,omitempty"`
	LastAuditDate    string    `json:"lastAuditDate,omitempty"`
	UtilizationRate  float64   `json:"utilizationRate,omitempty"`
	WarrantyExpiry   string    `json:"warrantyExpiry,omitempty"`
	AMCExpiry        string    `json:"amcExpiry,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type Store struct {
	db *bolt.DB
}

func NewStore(db *bolt.DB) (*Store, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(WorldStateBucketName))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// Get retrieves an asset by its ID from the worldstate bucket.
func (s *Store) Get(assetID string) (Asset, error) {
	var asset Asset
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorldStateBucketName))
		if b == nil {
			return fmt.Errorf("worldstate bucket not found")
		}
		val := b.Get([]byte(assetID))
		if val == nil {
			return fmt.Errorf("asset %s not found", assetID)
		}
		return json.Unmarshal(val, &asset)
	})
	return asset, err
}

// Put saves or updates an asset in the worldstate bucket.
func (s *Store) Put(asset Asset) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorldStateBucketName))
		if b == nil {
			return fmt.Errorf("worldstate bucket not found")
		}
		asset.UpdatedAt = time.Now().UTC()
		val, err := json.Marshal(asset)
		if err != nil {
			return err
		}
		return b.Put([]byte(asset.AssetID), val)
	})
}

// Delete removes an asset from the worldstate bucket.
func (s *Store) Delete(assetID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorldStateBucketName))
		if b == nil {
			return fmt.Errorf("worldstate bucket not found")
		}
		return b.Delete([]byte(assetID))
	})
}

// List retrieves all assets, optionally filtered by department.
func (s *Store) List(deptFilter string) ([]Asset, error) {
	var assets []Asset
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorldStateBucketName))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var asset Asset
			if err := json.Unmarshal(v, &asset); err != nil {
				return err
			}
			if deptFilter == "" || asset.DeptID == deptFilter {
				assets = append(assets, asset)
			}
			return nil
		})
	})
	return assets, err
}
