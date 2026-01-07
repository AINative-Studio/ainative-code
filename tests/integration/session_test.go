// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/AINative-studio/ainative-code/internal/session"
	"github.com/AINative-studio/ainative-code/tests/integration/fixtures"
	"github.com/AINative-studio/ainative-code/tests/integration/helpers"
	"github.com/stretchr/testify/suite"
)

// SessionIntegrationTestSuite tests session persistence and resume functionality.
type SessionIntegrationTestSuite struct {
	suite.Suite
	manager session.Manager
	cleanup func()
}

// SetupTest runs before each test in the suite.
func (s *SessionIntegrationTestSuite) SetupTest() {
	db, cleanup := helpers.SetupInMemoryDB(s.T())
	s.cleanup = cleanup
	s.manager = session.NewSQLiteManager(db)
}

// TearDownTest runs after each test in the suite.
func (s *SessionIntegrationTestSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// TestSessionCreationAndRetrieval tests creating a session and retrieving it.
func (s *SessionIntegrationTestSuite) TestSessionCreationAndRetrieval() {
	// Given: A new test session
	ctx := context.Background()
	testSession := fixtures.NewTestSession()

	// When: Creating the session
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err, "Failed to create session")

	// Then: The session should be retrievable
	retrieved, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err, "Failed to retrieve session")
	s.Equal(testSession.ID, retrieved.ID)
	s.Equal(testSession.Name, retrieved.Name)
	s.Equal(testSession.Status, retrieved.Status)
}

// TestSessionPersistenceAndResume tests session persistence across operations.
func (s *SessionIntegrationTestSuite) TestSessionPersistenceAndResume() {
	// Given: A session with multiple messages
	ctx := context.Background()
	testSession := fixtures.NewTestSessionWithName("Persistent Session")

	// Create session
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err, "Failed to create session")

	// Add multiple messages
	messages := fixtures.NewTestMessageThread(testSession.ID, 5)
	for _, msg := range messages {
		err := s.manager.AddMessage(ctx, msg)
		s.Require().NoError(err, "Failed to add message")
	}

	// When: Retrieving the session and its messages
	retrievedSession, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err, "Failed to retrieve session")

	retrievedMessages, err := s.manager.GetMessages(ctx, testSession.ID)
	s.Require().NoError(err, "Failed to retrieve messages")

	// Then: All data should be persisted correctly
	s.Equal(testSession.ID, retrievedSession.ID)
	s.Equal(testSession.Name, retrievedSession.Name)
	s.Len(retrievedMessages, 5, "Expected 5 messages")

	// Verify message thread structure
	for i, msg := range retrievedMessages {
		if i == 0 {
			s.Nil(msg.ParentID, "First message should have no parent")
		} else {
			s.NotNil(msg.ParentID, "Message should have parent")
			s.Equal(retrievedMessages[i-1].ID, *msg.ParentID, "Parent ID should match previous message")
		}
	}
}

// TestSessionUpdate tests updating session properties.
func (s *SessionIntegrationTestSuite) TestSessionUpdate() {
	// Given: An existing session
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	// When: Updating the session
	originalUpdatedAt := testSession.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference
	testSession.Name = "Updated Session Name"
	newModel := "claude-3-opus-20240229"
	testSession.Model = &newModel

	err = s.manager.UpdateSession(ctx, testSession)
	s.Require().NoError(err, "Failed to update session")

	// Then: The updates should be persisted
	retrieved, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err)
	s.Equal("Updated Session Name", retrieved.Name)
	s.Equal(newModel, *retrieved.Model)
	s.True(retrieved.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be more recent")
}

// TestSessionListingAndFiltering tests listing sessions with filters.
func (s *SessionIntegrationTestSuite) TestSessionListingAndFiltering() {
	// Given: Multiple sessions with different statuses
	ctx := context.Background()

	activeSessions := []*session.Session{
		fixtures.NewTestSessionWithName("Active 1"),
		fixtures.NewTestSessionWithName("Active 2"),
	}

	archivedSession := fixtures.NewTestSessionWithName("Archived")
	archivedSession.Status = session.StatusArchived

	// Create all sessions
	for _, sess := range activeSessions {
		err := s.manager.CreateSession(ctx, sess)
		s.Require().NoError(err)
	}
	err := s.manager.CreateSession(ctx, archivedSession)
	s.Require().NoError(err)

	// When: Listing active sessions only
	activeList, err := s.manager.ListSessions(ctx,
		session.WithStatus(session.StatusActive),
		session.WithLimit(10),
	)
	s.Require().NoError(err)

	// Then: Only active sessions should be returned
	s.Len(activeList, 2, "Expected 2 active sessions")
	for _, sess := range activeList {
		s.Equal(session.StatusActive, sess.Status)
	}

	// When: Listing all sessions
	allList, err := s.manager.ListSessions(ctx, session.WithLimit(10))
	s.Require().NoError(err)

	// Then: All sessions should be returned
	s.Len(allList, 3, "Expected 3 total sessions")
}

// TestSessionDeletion tests soft and hard deletion.
func (s *SessionIntegrationTestSuite) TestSessionDeletion() {
	// Given: A session with messages
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	msg := fixtures.NewTestMessage(testSession.ID, session.RoleUser, "Test message")
	err = s.manager.AddMessage(ctx, msg)
	s.Require().NoError(err)

	// When: Soft deleting the session
	err = s.manager.DeleteSession(ctx, testSession.ID)
	s.Require().NoError(err)

	// Then: Session should be marked as deleted
	retrieved, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err)
	s.Equal(session.StatusDeleted, retrieved.Status)

	// When: Hard deleting the session
	err = s.manager.HardDeleteSession(ctx, testSession.ID)
	s.Require().NoError(err)

	// Then: Session should not exist
	_, err = s.manager.GetSession(ctx, testSession.ID)
	s.Error(err, "Expected error when retrieving hard-deleted session")
	s.ErrorIs(err, session.ErrSessionNotFound)
}

// TestMessageOperations tests message CRUD operations.
func (s *SessionIntegrationTestSuite) TestMessageOperations() {
	// Given: A session
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	// When: Adding a message
	msg := fixtures.NewTestMessage(testSession.ID, session.RoleUser, "Hello, world!")
	err = s.manager.AddMessage(ctx, msg)
	s.Require().NoError(err)

	// Then: Message should be retrievable
	retrieved, err := s.manager.GetMessage(ctx, msg.ID)
	s.Require().NoError(err)
	s.Equal(msg.ID, retrieved.ID)
	s.Equal(msg.Content, retrieved.Content)
	s.Equal(msg.Role, retrieved.Role)

	// When: Updating the message
	msg.Content = "Updated content"
	err = s.manager.UpdateMessage(ctx, msg)
	s.Require().NoError(err)

	// Then: Update should be persisted
	retrieved, err = s.manager.GetMessage(ctx, msg.ID)
	s.Require().NoError(err)
	s.Equal("Updated content", retrieved.Content)

	// When: Deleting the message
	err = s.manager.DeleteMessage(ctx, msg.ID)
	s.Require().NoError(err)

	// Then: Message should not exist
	_, err = s.manager.GetMessage(ctx, msg.ID)
	s.Error(err, "Expected error when retrieving deleted message")
}

// TestSessionExportImport tests session export and import functionality.
func (s *SessionIntegrationTestSuite) TestSessionExportImport() {
	// Given: A session with messages
	ctx := context.Background()
	originalSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, originalSession)
	s.Require().NoError(err)

	messages := fixtures.NewTestMessageThread(originalSession.ID, 3)
	for _, msg := range messages {
		err := s.manager.AddMessage(ctx, msg)
		s.Require().NoError(err)
	}

	// When: Exporting the session
	var exportBuffer []byte
	exportWriter := &bytesWriter{data: &exportBuffer}
	err = s.manager.ExportSession(ctx, originalSession.ID, session.ExportFormatJSON, exportWriter)
	s.Require().NoError(err)
	s.NotEmpty(exportBuffer, "Export should produce data")

	// Delete the original session
	err = s.manager.HardDeleteSession(ctx, originalSession.ID)
	s.Require().NoError(err)

	// When: Importing the session
	exportReader := &bytesReader{data: exportBuffer}
	imported, err := s.manager.ImportSession(ctx, exportReader)
	s.Require().NoError(err)
	s.Equal(originalSession.ID, imported.ID)

	// Then: Imported session should match original
	importedMessages, err := s.manager.GetMessages(ctx, imported.ID)
	s.Require().NoError(err)
	s.Len(importedMessages, 3, "Expected 3 messages after import")
}

// TestSessionTokenTracking tests token usage tracking.
func (s *SessionIntegrationTestSuite) TestSessionTokenTracking() {
	// Given: A session with messages that have token counts
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	// Add messages with varying token counts
	tokenCounts := []int64{100, 200, 150}
	for _, tokens := range tokenCounts {
		msg := fixtures.NewTestMessage(testSession.ID, session.RoleUser, "Test")
		msg.TokensUsed = &tokens
		err := s.manager.AddMessage(ctx, msg)
		s.Require().NoError(err)
	}

	// When: Getting total tokens used
	totalTokens, err := s.manager.GetTotalTokensUsed(ctx, testSession.ID)
	s.Require().NoError(err)

	// Then: Total should match sum of individual messages
	expectedTotal := int64(450)
	s.Equal(expectedTotal, totalTokens, "Total tokens should be sum of all messages")

	// When: Getting session summary
	summary, err := s.manager.GetSessionSummary(ctx, testSession.ID)
	s.Require().NoError(err)

	// Then: Summary should include correct counts
	s.Equal(int64(3), summary.MessageCount)
	s.Equal(expectedTotal, summary.TotalTokens)
}

// TestSessionSearchFunctionality tests session and message search.
func (s *SessionIntegrationTestSuite) TestSessionSearchFunctionality() {
	// Given: Multiple sessions with searchable names
	ctx := context.Background()

	sessions := []*session.Session{
		fixtures.NewTestSessionWithName("OAuth Integration"),
		fixtures.NewTestSessionWithName("Design Tokens"),
		fixtures.NewTestSessionWithName("Integration Testing"),
	}

	for _, sess := range sessions {
		err := s.manager.CreateSession(ctx, sess)
		s.Require().NoError(err)
	}

	// When: Searching for sessions containing "Integration"
	results, err := s.manager.SearchSessions(ctx, "Integration", session.WithSearchLimit(10))
	s.Require().NoError(err)

	// Then: Both matching sessions should be found
	s.Len(results, 2, "Expected 2 sessions containing 'Integration'")

	// Verify both results contain "Integration" in name
	for _, result := range results {
		s.Contains(result.Name, "Integration")
	}
}

// TestConversationThreadRetrieval tests retrieving message threads.
func (s *SessionIntegrationTestSuite) TestConversationThreadRetrieval() {
	// Given: A session with a message thread
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	messages := fixtures.NewTestMessageThread(testSession.ID, 5)
	for _, msg := range messages {
		err := s.manager.AddMessage(ctx, msg)
		s.Require().NoError(err)
	}

	// When: Getting the conversation thread from the last message
	lastMessageID := messages[len(messages)-1].ID
	thread, err := s.manager.GetConversationThread(ctx, lastMessageID)
	s.Require().NoError(err)

	// Then: Should get all messages in the thread
	s.Len(thread, 5, "Expected full thread of 5 messages")

	// Verify thread order
	for i := 0; i < len(thread)-1; i++ {
		s.True(thread[i].Timestamp.Before(thread[i+1].Timestamp) ||
			thread[i].Timestamp.Equal(thread[i+1].Timestamp),
			"Thread should be ordered by timestamp")
	}
}

// TestSessionArchiveAndRestore tests archiving and restoring sessions.
func (s *SessionIntegrationTestSuite) TestSessionArchiveAndRestore() {
	// Given: An active session
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	// When: Archiving the session
	err = s.manager.ArchiveSession(ctx, testSession.ID)
	s.Require().NoError(err)

	// Then: Session should be archived
	retrieved, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err)
	s.Equal(session.StatusArchived, retrieved.Status)

	// When: Restoring the session (update status back to active)
	retrieved.Status = session.StatusActive
	err = s.manager.UpdateSession(ctx, retrieved)
	s.Require().NoError(err)

	// Then: Session should be active again
	restored, err := s.manager.GetSession(ctx, testSession.ID)
	s.Require().NoError(err)
	s.Equal(session.StatusActive, restored.Status)
}

// TestConcurrentSessionOperations tests thread-safety of session operations.
func (s *SessionIntegrationTestSuite) TestConcurrentSessionOperations() {
	// Given: A session
	ctx := context.Background()
	testSession := fixtures.NewTestSession()
	err := s.manager.CreateSession(ctx, testSession)
	s.Require().NoError(err)

	// When: Adding messages concurrently
	concurrentOps := 10
	done := make(chan bool, concurrentOps)
	errors := make(chan error, concurrentOps)

	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			msg := fixtures.NewTestMessage(testSession.ID, session.RoleUser, "Concurrent message")
			if err := s.manager.AddMessage(ctx, msg); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		<-done
	}
	close(errors)

	// Then: All operations should succeed
	s.Empty(errors, "No errors should occur during concurrent operations")

	// Verify all messages were added
	messages, err := s.manager.GetMessages(ctx, testSession.ID)
	s.Require().NoError(err)
	s.Len(messages, concurrentOps, "All concurrent messages should be persisted")
}

// Helper types for export/import testing
type bytesWriter struct {
	data *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

type bytesReader struct {
	data   []byte
	offset int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// TestSessionIntegrationTestSuite runs the test suite.
func TestSessionIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationTestSuite))
}
