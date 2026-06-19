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
	FlagPositionMiddleLeft   FlagPosition = "front_left"
	FlagPositionMiddleCentre FlagPosition = "front_centre"
	FlagPositionMiddleRight  FlagPosition = "front_right"
	FlagPositionBackLeft     FlagPosition = "front_left"
	FlagPositionBackCentre   FlagPosition = "front_centre"
	FlagPositionBackRight    FlagPosition = "front_right"
)
