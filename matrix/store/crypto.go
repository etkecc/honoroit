package store

import (
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// NOTE: functions in that file are for crypto.Store implementation
// ref: https://pkg.go.dev/maunium.net/go/mautrix/crypto#Store

// Flush does nothing for this implementation as data is already persisted in the database.
func (s *Store) Flush() error {
	return nil
}

// PutNextBatch stores the next sync batch token for the current account.
func (s *Store) PutNextBatch(nextBatch string) {
	s.s.PutNextBatch(nextBatch)
}

// GetNextBatch retrieves the next sync batch token for the current account.
func (s *Store) GetNextBatch() string {
	return s.s.GetNextBatch()
}

// PutAccount stores an OlmAccount in the database.
func (s *Store) PutAccount(account *crypto.OlmAccount) error {
	return s.s.PutAccount(account)
}

// GetAccount retrieves an OlmAccount from the database.
func (s *Store) GetAccount() (*crypto.OlmAccount, error) {
	return s.s.GetAccount()
}

// HasSession returns whether there is an Olm session for the given sender key.
func (s *Store) HasSession(key id.SenderKey) bool {
	return s.s.HasSession(key)
}

// GetSessions returns all the known Olm sessions for a sender key.
func (s *Store) GetSessions(key id.SenderKey) (crypto.OlmSessionList, error) {
	return s.s.GetSessions(key)
}

// GetLatestSession retrieves the Olm session for a given sender key from the database that has the largest ID.
func (s *Store) GetLatestSession(key id.SenderKey) (*crypto.OlmSession, error) {
	return s.s.GetLatestSession(key)
}

// AddSession persists an Olm session for a sender in the database.
func (s *Store) AddSession(key id.SenderKey, session *crypto.OlmSession) error {
	return s.s.AddSession(key, session)
}

// UpdateSession replaces the Olm session for a sender in the database.
func (s *Store) UpdateSession(key id.SenderKey, session *crypto.OlmSession) error {
	return s.s.UpdateSession(key, session)
}

// PutGroupSession stores an inbound Megolm group session for a room, sender and session.
func (s *Store) PutGroupSession(roomID id.RoomID, senderKey id.SenderKey, sessionID id.SessionID, session *crypto.InboundGroupSession) error {
	return s.s.PutGroupSession(roomID, senderKey, sessionID, session)
}

// GetGroupSession retrieves an inbound Megolm group session for a room, sender and session.
func (s *Store) GetGroupSession(roomID id.RoomID, senderKey id.SenderKey, sessionID id.SessionID) (*crypto.InboundGroupSession, error) {
	return s.s.GetGroupSession(roomID, senderKey, sessionID)
}

// PutWithheldGroupSession tells the store that a specific Megolm session was withheld.
// nolint // method is part of interface and cannot be changed
func (s *Store) PutWithheldGroupSession(content event.RoomKeyWithheldEventContent) error {
	return s.s.PutWithheldGroupSession(content)
}

// GetWithheldGroupSession gets the event content that was previously inserted with PutWithheldGroupSession.
func (s *Store) GetWithheldGroupSession(roomID id.RoomID, senderKey id.SenderKey, sessionID id.SessionID) (*event.RoomKeyWithheldEventContent, error) {
	return s.s.GetWithheldGroupSession(roomID, senderKey, sessionID)
}

// GetGroupSessionsForRoom gets all the inbound Megolm sessions for a specific room. This is used for creating key
// export files. Unlike GetGroupSession, this should not return any errors about withheld keys.
func (s *Store) GetGroupSessionsForRoom(roomID id.RoomID) ([]*crypto.InboundGroupSession, error) {
	return s.s.GetGroupSessionsForRoom(roomID)
}

// GetGroupSessionsForRoom gets all the inbound Megolm sessions in the store. This is used for creating key export
// files. Unlike GetGroupSession, this should not return any errors about withheld keys.
func (s *Store) GetAllGroupSessions() ([]*crypto.InboundGroupSession, error) {
	return s.s.GetAllGroupSessions()
}

// AddOutboundGroupSession stores an outbound Megolm session, along with the information about the room and involved devices.
func (s *Store) AddOutboundGroupSession(session *crypto.OutboundGroupSession) (err error) {
	return s.s.AddOutboundGroupSession(session)
}

// UpdateOutboundGroupSession replaces an outbound Megolm session with for same room and session ID.
func (s *Store) UpdateOutboundGroupSession(session *crypto.OutboundGroupSession) error {
	return s.s.UpdateOutboundGroupSession(session)
}

// GetOutboundGroupSession retrieves the outbound Megolm session for the given room ID.
func (s *Store) GetOutboundGroupSession(roomID id.RoomID) (*crypto.OutboundGroupSession, error) {
	return s.s.GetOutboundGroupSession(roomID)
}

// RemoveOutboundGroupSession removes the outbound Megolm session for the given room ID.
func (s *Store) RemoveOutboundGroupSession(roomID id.RoomID) error {
	return s.s.RemoveOutboundGroupSession(roomID)
}

// ValidateMessageIndex returns whether the given event information match the ones stored in the database
// for the given sender key, session ID and index.
// If the event information was not yet stored, it's stored now.
func (s *Store) ValidateMessageIndex(senderKey id.SenderKey, sessionID id.SessionID, eventID id.EventID, index uint, timestamp int64) bool {
	return s.s.ValidateMessageIndex(senderKey, sessionID, eventID, index, timestamp)
}

// GetDevices returns a map of device IDs to device identities, including the identity and signing keys, for a given user ID.
func (s *Store) GetDevices(userID id.UserID) (map[id.DeviceID]*crypto.DeviceIdentity, error) {
	return s.s.GetDevices(userID)
}

// GetDevice returns the device dentity for a given user and device ID.
func (s *Store) GetDevice(userID id.UserID, deviceID id.DeviceID) (*crypto.DeviceIdentity, error) {
	return s.s.GetDevice(userID, deviceID)
}

// FindDeviceByKey finds a specific device by its sender key.
func (s *Store) FindDeviceByKey(userID id.UserID, identityKey id.IdentityKey) (*crypto.DeviceIdentity, error) {
	return s.s.FindDeviceByKey(userID, identityKey)
}

// PutDevice stores a single device for a user, replacing it if it exists already.
func (s *Store) PutDevice(userID id.UserID, device *crypto.DeviceIdentity) error {
	return s.s.PutDevice(userID, device)
}

// PutDevices stores the device identity information for the given user ID.
func (s *Store) PutDevices(userID id.UserID, devices map[id.DeviceID]*crypto.DeviceIdentity) error {
	return s.s.PutDevices(userID, devices)
}

// FilterTrackedUsers finds all of the user IDs out of the given ones for which the database contains identity information.
func (s *Store) FilterTrackedUsers(users []id.UserID) []id.UserID {
	return s.s.FilterTrackedUsers(users)
}

// PutCrossSigningKey stores a cross-signing key of some user along with its usage.
func (s *Store) PutCrossSigningKey(userID id.UserID, usage id.CrossSigningUsage, key id.Ed25519) error {
	return s.s.PutCrossSigningKey(userID, usage, key)
}

// GetCrossSigningKeys retrieves a user's stored cross-signing keys.
func (s *Store) GetCrossSigningKeys(userID id.UserID) (map[id.CrossSigningUsage]id.Ed25519, error) {
	return s.s.GetCrossSigningKeys(userID)
}

// PutSignature stores a signature of a cross-signing or device key along with the signer's user ID and key.
func (s *Store) PutSignature(signedUserID id.UserID, signedKey id.Ed25519, signerUserID id.UserID, signerKey id.Ed25519, signature string) error {
	return s.s.PutSignature(signedUserID, signedKey, signerUserID, signerKey, signature)
}

// GetSignaturesForKeyBy retrieves the stored signatures for a given cross-signing or device key, by the given signer.
func (s *Store) GetSignaturesForKeyBy(userID id.UserID, key id.Ed25519, signerID id.UserID) (map[id.Ed25519]string, error) {
	return s.s.GetSignaturesForKeyBy(userID, key, signerID)
}

// IsKeySignedBy returns whether a cross-signing or device key is signed by the given signer.
func (s *Store) IsKeySignedBy(userID id.UserID, key id.Ed25519, signerID id.UserID, signerKey id.Ed25519) (bool, error) {
	return s.s.IsKeySignedBy(userID, key, signerID, signerKey)
}

// DropSignaturesByKey deletes the signatures made by the given user and key from the store. It returns the number of signatures deleted.
func (s *Store) DropSignaturesByKey(userID id.UserID, key id.Ed25519) (int64, error) {
	return s.s.DropSignaturesByKey(userID, key)
}
