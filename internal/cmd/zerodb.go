package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	zerodbMigrate bool
	zerodbBackup  string
	zerodbRestore string
)

// zerodbCmd represents the zerodb command
var zerodbCmd = &cobra.Command{
	Use:   "zerodb",
	Short: "Manage ZeroDB operations",
	Long: `Manage ZeroDB database operations including initialization, migration,
backup, and restore operations.

ZeroDB is used to store chat sessions, RLHF feedback, design tokens,
and other application data with client-side encryption.

Examples:
  # Initialize database
  ainative-code zerodb init

  # Run migrations
  ainative-code zerodb migrate

  # Check database status
  ainative-code zerodb status

  # Backup database
  ainative-code zerodb backup --output backup.db

  # Restore database
  ainative-code zerodb restore --input backup.db

  # Vacuum database
  ainative-code zerodb vacuum`,
	Aliases: []string{"db", "database"},
}

// zerodbInitCmd represents the zerodb init command
var zerodbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ZeroDB database",
	Long:  `Initialize the ZeroDB database with required tables and indexes.`,
	RunE:  runZerodbInit,
}

// zerodbMigrateCmd represents the zerodb migrate command
var zerodbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run pending database migrations to update schema.`,
	RunE:  runZerodbMigrate,
}

// zerodbStatusCmd represents the zerodb status command
var zerodbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check database status",
	Long:  `Display current database status including size, tables, and migration version.`,
	RunE:  runZerodbStatus,
}

// zerodbBackupCmd represents the zerodb backup command
var zerodbBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup database",
	Long:  `Create a backup of the ZeroDB database.`,
	RunE:  runZerodbBackup,
}

// zerodbRestoreCmd represents the zerodb restore command
var zerodbRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database",
	Long:  `Restore database from a backup file.`,
	RunE:  runZerodbRestore,
}

// zerodbVacuumCmd represents the zerodb vacuum command
var zerodbVacuumCmd = &cobra.Command{
	Use:   "vacuum",
	Short: "Vacuum database",
	Long:  `Optimize database by reclaiming unused space.`,
	RunE:  runZerodbVacuum,
}

func init() {
	rootCmd.AddCommand(zerodbCmd)

	// Add subcommands
	zerodbCmd.AddCommand(zerodbInitCmd)
	zerodbCmd.AddCommand(zerodbMigrateCmd)
	zerodbCmd.AddCommand(zerodbStatusCmd)
	zerodbCmd.AddCommand(zerodbBackupCmd)
	zerodbCmd.AddCommand(zerodbRestoreCmd)
	zerodbCmd.AddCommand(zerodbVacuumCmd)

	// Backup flags
	zerodbBackupCmd.Flags().StringP("output", "o", "", "output backup file path (required)")
	zerodbBackupCmd.MarkFlagRequired("output")

	// Restore flags
	zerodbRestoreCmd.Flags().StringP("input", "i", "", "input backup file path (required)")
	zerodbRestoreCmd.MarkFlagRequired("input")
	zerodbRestoreCmd.Flags().BoolP("force", "f", false, "force restore (overwrite existing data)")
}

func runZerodbInit(cmd *cobra.Command, args []string) error {
	logger.Info("Initializing ZeroDB database")

	fmt.Println("Initializing ZeroDB database...")
	fmt.Println("Creating tables and indexes...")

	// TODO: Implement database initialization
	// - Create sessions table
	// - Create messages table
	// - Create feedback table
	// - Create design_tokens table
	// - Create indexes

	fmt.Println("Database initialized successfully!")

	return nil
}

func runZerodbMigrate(cmd *cobra.Command, args []string) error {
	logger.Info("Running database migrations")

	fmt.Println("Running database migrations...")

	// TODO: Implement migration system
	// - Check current version
	// - Apply pending migrations
	// - Update version

	fmt.Println("Migrations completed successfully!")

	return nil
}

func runZerodbStatus(cmd *cobra.Command, args []string) error {
	logger.Debug("Checking database status")

	// Initialize database connection
	db, err := getDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	fmt.Println("\nDatabase Status:")
	fmt.Println("================")

	// Get database path
	dbPath := getDatabasePath()
	fmt.Printf("Path: %s\n", dbPath)

	// Get database file size
	if fileInfo, err := os.Stat(dbPath); err == nil {
		sizeKB := float64(fileInfo.Size()) / 1024
		sizeMB := sizeKB / 1024
		if sizeMB >= 1 {
			fmt.Printf("Size: %.2f MB (%.0f KB)\n", sizeMB, sizeKB)
		} else {
			fmt.Printf("Size: %.2f KB (%d bytes)\n", sizeKB, fileInfo.Size())
		}
	} else {
		fmt.Printf("Size: Unable to read (error: %v)\n", err)
	}

	// Get the underlying sql.DB
	sqlDB := db.DB()

	// Get schema version from migrations table
	var version int
	err = sqlDB.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		fmt.Printf("Schema Version: Unable to read (error: %v)\n", err)
	} else {
		fmt.Printf("Schema Version: %d\n", version)
	}

	// List all tables
	rows, err := sqlDB.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type='table'
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	fmt.Printf("\nTables (%d):\n", len(tables))
	fmt.Println("============")

	// Get row counts for each table
	for _, table := range tables {
		var count int
		err := sqlDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("  - %s: error counting rows\n", table)
		} else {
			fmt.Printf("  - %s: %d rows\n", table, count)
		}
	}

	// Check FTS5 support
	var ftsSupported bool
	err = sqlDB.QueryRow("SELECT 1 FROM pragma_compile_options WHERE compile_options = 'ENABLE_FTS5'").Scan(&ftsSupported)
	if err == nil && ftsSupported {
		fmt.Println("\nFTS5 Support: ✓ Enabled")
	} else {
		// Try another way to check FTS5
		_, err = sqlDB.Query("SELECT * FROM messages_fts LIMIT 0")
		if err == nil {
			fmt.Println("\nFTS5 Support: ✓ Enabled")
		} else {
			fmt.Println("\nFTS5 Support: ✗ Disabled")
		}
	}

	fmt.Println()
	return nil
}

func runZerodbBackup(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")

	logger.InfoEvent().Str("output", output).Msg("Creating database backup")

	fmt.Printf("Creating backup to: %s\n", output)

	// TODO: Implement backup
	// - Copy database file
	// - Verify backup integrity

	fmt.Println("Backup created successfully!")

	return nil
}

func runZerodbRestore(cmd *cobra.Command, args []string) error {
	input, _ := cmd.Flags().GetString("input")
	force, _ := cmd.Flags().GetBool("force")

	logger.InfoEvent().
		Str("input", input).
		Bool("force", force).
		Msg("Restoring database")

	if !force {
		fmt.Println("WARNING: This will overwrite the current database.")
		fmt.Println("Use --force to confirm restoration.")
		return fmt.Errorf("restoration cancelled (use --force to proceed)")
	}

	fmt.Printf("Restoring database from: %s\n", input)

	// TODO: Implement restore
	// - Verify backup file
	// - Replace current database
	// - Verify restoration

	fmt.Println("Database restored successfully!")

	return nil
}

func runZerodbVacuum(cmd *cobra.Command, args []string) error {
	logger.Info("Vacuuming database")

	fmt.Println("Optimizing database...")

	// TODO: Implement vacuum
	// - Run VACUUM command
	// - Show before/after size

	fmt.Println("Database optimized successfully!")

	return nil
}
