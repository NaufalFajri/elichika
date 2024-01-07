package gamedata

import (
	"elichika/dictionary"
	"elichika/enum"
	"elichika/model"
	"elichika/utils"

	"fmt"

	"xorm.io/xorm"
)

type TowerFloor struct {
	// from m_tower_composition
	TowerID int `xorm:"pk 'tower_id'"`
	FloorNo int `xorm:"pk 'floor_no'"`
	// Name DictionaryString `xorm:"'name'"`
	// ThumbnailAssetPath *string `xorm:"'thumbnail_asset_path'"`
	// PopUpThumbnailAssetPath string `xorm:"'popup_thumbnail_asset_path'"`
	ConsumePerformance bool `xorm:"'consume_performance'"`
	TowerCellType      int  `xorm:"'tower_cell_type'"`
	// ScenarioScriptAssetPath *string `xorm:"'scenario_script_asset_path'"`
	// LiveDifficultyID int `xorm:"'live_difficulty_id'"`
	TargetVoltage int `xorm:"'target_voltage'"`
	// SuperStageAssetPath *string `xorm:"'super_stage_asset_path'"`
	// StillAssetPath *string `xorm:"'still_asset_path'"`
	// MusicID *int  `xorm:"'music_id'"`
	TowerClearRewardID    *int            `xorm:"'tower_clear_reward_id'"`
	TowerClearRewards     []model.Content `xorm:"-"` // from: m_tower_clear_reward
	TowerProgressRewardID *int            `xorm:"'tower_progress_reward_id'"`
	TowerProgressRewards  []model.Content `xorm:"-"` // from: m_tower_progress_reward
}

func (tf *TowerFloor) populate(gamedata *Gamedata, masterdata_db, serverdata_db *xorm.Session, dictionary *dictionary.Dictionary) {
	if tf.TowerClearRewardID != nil {
		err := masterdata_db.Table("m_tower_clear_reward").Where("tower_clear_reward_id = ?", *tf.TowerClearRewardID).
			Find(&tf.TowerClearRewards)
		utils.CheckErr(err)
	}
	if tf.TowerProgressRewardID != nil {
		err := masterdata_db.Table("m_tower_progress_reward").Where("tower_progress_reward_id = ?", *tf.TowerProgressRewardID).
			Find(&tf.TowerProgressRewards)
		utils.CheckErr(err)
	}
}

type Tower struct {
	// from m_tower
	TowerID int `xorm:"pk 'tower_id'"`
	// Title DictionaryString `xorm:"'title'"`
	// ThumbnailAssetPath string `xorm:"'thumbnail_asset_path'"`
	// DisplayOrder int `xorm:"'display_order'"`
	TowerCompositionID   int          `xorm:"'tower_composition_id'"`
	Floor                []TowerFloor `xorm:"-"` // from m_tower_composition, 1 indexed
	FloorCount           int          `xorm:"-"`
	IsVoltageRanked      bool         `xorm:"-"`
	TradeMasterID        int          `xorm:"'trade_master_id'"`
	EntryRestrictionType int          `xorm:"'entry_restriction_type'"`
	// EntryRestrictionCondition *int `xorm:"'entry_restriction_condition'"`
	CardUseLimit      int `xorm:"'card_use_limit'"`
	CardRecoveryLimit int `xorm:"'card_recovery_limit'"`
	// FreeRecoveryPointAt int `xorm:"'free_recover_point_recovery_at'"`
	// FreeRecoveryPointMaxCount int `xorm:"'free_recover_point_max_count'"`
	RecoverCostBySnsCoin int `xorm:"'recover_cost_by_sns_coin'"`
	// BackgroundAssetPath string `xorm:"'background_asset_path'"`
}

func (t *Tower) populate(gamedata *Gamedata, masterdata_db, serverdata_db *xorm.Session, dictionary *dictionary.Dictionary) {
	err := masterdata_db.Table("m_tower_composition").Where("tower_id = ?", t.TowerID).OrderBy("floor_no").Find(&t.Floor)
	utils.CheckErr(err)
	t.FloorCount = len(t.Floor)
	t.Floor = append([]TowerFloor{TowerFloor{}}, t.Floor...)
	for i := range t.Floor {
		t.Floor[i].populate(gamedata, masterdata_db, serverdata_db, dictionary)
		t.IsVoltageRanked = t.IsVoltageRanked || (t.Floor[i].TowerCellType == enum.TowerCellTypeBonusLive)
	}
}

func loadTower(gamedata *Gamedata, masterdata_db, serverdata_db *xorm.Session, dictionary *dictionary.Dictionary) {
	fmt.Println("Loading Tower")
	gamedata.Tower = make(map[int]*Tower)
	err := masterdata_db.Table("m_tower").Find(&gamedata.Tower)
	utils.CheckErr(err)
	for _, tower := range gamedata.Tower {
		tower.populate(gamedata, masterdata_db, serverdata_db, dictionary)
	}
}
func init() {
	addLoadFunc(loadTower)
}