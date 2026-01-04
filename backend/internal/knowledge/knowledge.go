package knowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// =============================================================================
// Knowledge Base Service
// =============================================================================

// Service handles knowledge base operations
type Service struct {
	vectorStore VectorStore
	embedder    Embedder
	log         *logger.Logger
}

// NewService creates a new knowledge service
func NewService(vectorStore VectorStore, embedder Embedder, log *logger.Logger) *Service {
	return &Service{
		vectorStore: vectorStore,
		embedder:    embedder,
		log:         log,
	}
}

// VectorStore interface for vector database operations
type VectorStore interface {
	// Store stores chunks with embeddings
	StoreChunks(ctx context.Context, kbID uuid.UUID, chunks []Chunk) error

	// Search performs similarity search
	Search(ctx context.Context, kbID uuid.UUID, embedding []float32, limit int) ([]SearchResult, error)

	// Delete removes all chunks for a document
	DeleteDocument(ctx context.Context, documentID uuid.UUID) error

	// DeleteKnowledgeBase removes all data for a knowledge base
	DeleteKnowledgeBase(ctx context.Context, kbID uuid.UUID) error
}

// Embedder interface for generating embeddings
type Embedder interface {
	// Embed generates embeddings for text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the embedding dimension
	Dimension() int
}

// Chunk represents a text chunk with its embedding
type Chunk struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Content    string
	Embedding  []float32
	Metadata   map[string]interface{}
	Index      int
}

// SearchResult represents a search result
type SearchResult struct {
	ChunkID    uuid.UUID
	DocumentID uuid.UUID
	Content    string
	Score      float32
	Metadata   map[string]interface{}
}

// =============================================================================
// Document Ingestion
// =============================================================================

// IngestRequest represents a document ingestion request
type IngestRequest struct {
	KnowledgeBaseID uuid.UUID
	Source          string
	SourceType      string // file, url, text, repository
	Content         string
	Metadata        map[string]interface{}
}

// IngestResult represents the result of document ingestion
type IngestResult struct {
	DocumentID  uuid.UUID
	ChunkCount  int
	ContentHash string
	Duration    time.Duration
}

// Ingest ingests a document into the knowledge base
func (s *Service) Ingest(ctx context.Context, req *IngestRequest) (*IngestResult, error) {
	start := time.Now()

	// Generate content hash
	hash := sha256.Sum256([]byte(req.Content))
	contentHash := hex.EncodeToString(hash[:])

	// Create document record
	documentID := uuid.New()

	// Chunk the content
	chunks := s.chunkContent(req.Content, documentID)

	s.log.Infow("chunking complete", 
		"document_id", documentID, 
		"chunk_count", len(chunks),
	)

	// Generate embeddings for chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := s.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Attach embeddings to chunks
	for i := range chunks {
		chunks[i].Embedding = embeddings[i]
		chunks[i].Metadata = req.Metadata
	}

	// Store chunks
	if err := s.vectorStore.StoreChunks(ctx, req.KnowledgeBaseID, chunks); err != nil {
		return nil, fmt.Errorf("failed to store chunks: %w", err)
	}

	return &IngestResult{
		DocumentID:  documentID,
		ChunkCount:  len(chunks),
		ContentHash: contentHash,
		Duration:    time.Since(start),
	}, nil
}

// chunkContent splits content into chunks with overlap
func (s *Service) chunkContent(content string, documentID uuid.UUID) []Chunk {
	const (
		chunkSize   = 1000 // characters
		chunkOverlap = 200
	)

	// Split by paragraphs first
	paragraphs := strings.Split(content, "\n\n")
	
	var chunks []Chunk
	var currentChunk strings.Builder
	chunkIndex := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would exceed chunk size
		if currentChunk.Len()+len(para) > chunkSize && currentChunk.Len() > 0 {
			// Save current chunk
			chunks = append(chunks, Chunk{
				ID:         uuid.New(),
				DocumentID: documentID,
				Content:    currentChunk.String(),
				Index:      chunkIndex,
			})
			chunkIndex++

			// Start new chunk with overlap
			content := currentChunk.String()
			currentChunk.Reset()
			
			// Add overlap from end of previous chunk
			if len(content) > chunkOverlap {
				currentChunk.WriteString(content[len(content)-chunkOverlap:])
				currentChunk.WriteString(" ")
			}
		}

		currentChunk.WriteString(para)
		currentChunk.WriteString("\n\n")
	}

	// Don't forget the last chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			ID:         uuid.New(),
			DocumentID: documentID,
			Content:    strings.TrimSpace(currentChunk.String()),
			Index:      chunkIndex,
		})
	}

	return chunks
}

// =============================================================================
// Querying
// =============================================================================

// QueryRequest represents a query request
type QueryRequest struct {
	KnowledgeBaseIDs []uuid.UUID
	Query            string
	Limit            int
	MinScore         float32
}

// QueryResult represents query results
type QueryResult struct {
	Results  []SearchResult
	Duration time.Duration
}

// Query searches the knowledge base
func (s *Service) Query(ctx context.Context, req *QueryRequest) (*QueryResult, error) {
	start := time.Now()

	// Generate embedding for query
	embedding, err := s.embedder.Embed(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Search each knowledge base
	var allResults []SearchResult
	for _, kbID := range req.KnowledgeBaseIDs {
		results, err := s.vectorStore.Search(ctx, kbID, embedding, limit)
		if err != nil {
			s.log.Warnw("search failed for knowledge base", "kb_id", kbID, "error", err)
			continue
		}
		allResults = append(allResults, results...)
	}

	// Filter by minimum score
	if req.MinScore > 0 {
		filtered := make([]SearchResult, 0, len(allResults))
		for _, r := range allResults {
			if r.Score >= req.MinScore {
				filtered = append(filtered, r)
			}
		}
		allResults = filtered
	}

	// Sort by score and limit
	// (In production, use a proper sorting algorithm)
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return &QueryResult{
		Results:  allResults,
		Duration: time.Since(start),
	}, nil
}

// =============================================================================
// Repository Indexing
// =============================================================================

// RepositoryIndexer handles repository content indexing
type RepositoryIndexer struct {
	service *Service
	log     *logger.Logger
}

// NewRepositoryIndexer creates a new repository indexer
func NewRepositoryIndexer(service *Service, log *logger.Logger) *RepositoryIndexer {
	return &RepositoryIndexer{
		service: service,
		log:     log,
	}
}

// IndexRepositoryRequest represents a repository indexing request
type IndexRepositoryRequest struct {
	KnowledgeBaseID uuid.UUID
	Repository      *models.Repository
	Files           []RepositoryFile
}

// RepositoryFile represents a file in a repository
type RepositoryFile struct {
	Path     string
	Content  string
	Language string
}

// IndexRepository indexes a repository into the knowledge base
func (i *RepositoryIndexer) IndexRepository(ctx context.Context, req *IndexRepositoryRequest) error {
	i.log.Infow("indexing repository", 
		"repo", req.Repository.FullName, 
		"file_count", len(req.Files),
	)

	for _, file := range req.Files {
		// Skip binary or very large files
		if len(file.Content) > 100000 {
			continue
		}

		metadata := map[string]interface{}{
			"path":       file.Path,
			"language":   file.Language,
			"repository": req.Repository.FullName,
		}

		_, err := i.service.Ingest(ctx, &IngestRequest{
			KnowledgeBaseID: req.KnowledgeBaseID,
			Source:          file.Path,
			SourceType:      "repository",
			Content:         file.Content,
			Metadata:        metadata,
		})
		if err != nil {
			i.log.Warnw("failed to index file", "path", file.Path, "error", err)
			continue
		}
	}

	i.log.Infow("repository indexed", "repo", req.Repository.FullName)
	return nil
}

// =============================================================================
// Mock Implementations for Development
// =============================================================================

// MockVectorStore is a simple in-memory vector store for development
type MockVectorStore struct {
	chunks map[uuid.UUID][]Chunk
}

// NewMockVectorStore creates a new mock vector store
func NewMockVectorStore() *MockVectorStore {
	return &MockVectorStore{
		chunks: make(map[uuid.UUID][]Chunk),
	}
}

func (s *MockVectorStore) StoreChunks(ctx context.Context, kbID uuid.UUID, chunks []Chunk) error {
	s.chunks[kbID] = append(s.chunks[kbID], chunks...)
	return nil
}

func (s *MockVectorStore) Search(ctx context.Context, kbID uuid.UUID, embedding []float32, limit int) ([]SearchResult, error) {
	chunks := s.chunks[kbID]
	if len(chunks) == 0 {
		return nil, nil
	}

	// Simple mock search - return first N chunks
	var results []SearchResult
	for i, chunk := range chunks {
		if i >= limit {
			break
		}
		results = append(results, SearchResult{
			ChunkID:    chunk.ID,
			DocumentID: chunk.DocumentID,
			Content:    chunk.Content,
			Score:      0.9 - float32(i)*0.1,
			Metadata:   chunk.Metadata,
		})
	}

	return results, nil
}

func (s *MockVectorStore) DeleteDocument(ctx context.Context, documentID uuid.UUID) error {
	for kbID, chunks := range s.chunks {
		var filtered []Chunk
		for _, chunk := range chunks {
			if chunk.DocumentID != documentID {
				filtered = append(filtered, chunk)
			}
		}
		s.chunks[kbID] = filtered
	}
	return nil
}

func (s *MockVectorStore) DeleteKnowledgeBase(ctx context.Context, kbID uuid.UUID) error {
	delete(s.chunks, kbID)
	return nil
}

// MockEmbedder generates mock embeddings for development
type MockEmbedder struct {
	dimension int
}

// NewMockEmbedder creates a new mock embedder
func NewMockEmbedder(dimension int) *MockEmbedder {
	if dimension <= 0 {
		dimension = 1536 // OpenAI default
	}
	return &MockEmbedder{dimension: dimension}
}

func (e *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Generate a simple hash-based embedding for consistency
	embedding := make([]float32, e.dimension)
	hash := sha256.Sum256([]byte(text))
	for i := 0; i < e.dimension && i < 32; i++ {
		embedding[i] = float32(hash[i%32]) / 255.0
	}
	return embedding, nil
}

func (e *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := e.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (e *MockEmbedder) Dimension() int {
	return e.dimension
}

