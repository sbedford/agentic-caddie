package models

type Location string
type ShotType string
type ShotResult string
type StrikeQuality string
type CompetitionType string
type RoundType string
type FlagPosition string

const (
	CompetitionTypeStableford CompetitionType = "stableford"
	CompetitionTypeStroke     CompetitionType = "stroke"
	CompetitionTypeOther      CompetitionType = "other"

	RoundTypeSocial      RoundType = "social"
	RoundTypeCompetition RoundType = "competition"
	RoundTypePractice    RoundType = "practice"

	LocationTee           Location = "tee"
	LocationFairway       Location = "fairway"
	LocationRough         Location = "rough"
	LocationBunker        Location = "bunker"
	LocationGreen         Location = "green"
	LocationHazard        Location = "hazard"
	LocationHoleCompleted Location = "hole completed"
	LocationOutOfBounds   Location = "ob"
	LocationLostBall      Location = "lost"

	ShotTypeTee      ShotType = "tee"
	ShotTypeApproach ShotType = "approach"
	ShotTypeLayup    ShotType = "layup"
	ShotTypeChip     ShotType = "chip"
	ShotTypePitch    ShotType = "pitch"
	ShotTypeBunker   ShotType = "bunker"
	ShotTypeRecord   ShotType = "recovery"

	ShotResultLeft  ShotResult = "left"
	ShotResultRight ShotResult = "righ"
	ShotResultShort ShotResult = "short"
	ShotResultLong  ShotResult = "long"

	StrikeQualityClean StrikeQuality = "clean"
	StrikeQualityFat   StrikeQuality = "fat"
	StrikeQualityThin  StrikeQuality = "thin"
	StrikeQualityShank StrikeQuality = "shank"

	FlagPositionFrontLeft    FlagPosition = "front_left"
	FlagPositionFrontCentre  FlagPosition = "front_centre"
	FlagPositionFrontRight   FlagPosition = "front_right"
	FlagPositionMiddleLeft   FlagPosition = "middle_left"
	FlagPositionMiddleCentre FlagPosition = "middle_centre"
	FlagPositionMiddleRight  FlagPosition = "middle_right"
	FlagPositionBackLeft     FlagPosition = "back_left"
	FlagPositionBackCentre   FlagPosition = "back_centre"
	FlagPositionBackRight    FlagPosition = "back_right"
)
