package store

import (
	"versio-index/db"
	"versio-index/index/errors"
	"versio-index/index/model"

	"github.com/golang/protobuf/proto"
)

type RepoReadOnlyOperations interface {
	Snapshot() RepoReadOnlyOperations
	ReadRepo() (*model.Repo, error)
	ListWorkspace(branch string) ([]*model.WorkspaceEntry, error)
	ReadFromWorkspace(branch, path string) (*model.WorkspaceEntry, error)
	ReadBranch(branch string) (*model.Branch, error)
	ReadObject(addr string) (*model.Object, error)
	ReadCommit(addr string) (*model.Commit, error)
	ListTree(addr, from string, results int) ([]*model.Entry, error)
	ReadTreeEntry(treeAddress, name string, entryType model.Entry_Type) (*model.Entry, error)
}

type RepoOperations interface {
	RepoReadOnlyOperations
	WriteToWorkspacePath(branch, path string, entry *model.WorkspaceEntry) error
	ClearWorkspace(branch string)
	WriteTree(address string, entries []*model.Entry) error
	WriteObject(addr string, object *model.Object) error
	WriteCommit(addr string, commit *model.Commit) error
	WriteBranch(name string, branch *model.Branch) error
	DeleteBranch(name string)
	WriteRepo(repo *model.Repo) error
}

type KVRepoReadOnlyOperations struct {
	query db.ReadQuery
	store db.Store
}

type KVRepoOperations struct {
	*KVRepoReadOnlyOperations
	query db.Query
}

func (s *KVRepoReadOnlyOperations) Snapshot() RepoReadOnlyOperations {
	return &KVRepoReadOnlyOperations{
		query: s.query.Snapshot(),
		store: s.store,
	}
}

func (s *KVRepoReadOnlyOperations) ReadRepo() (*model.Repo, error) {
	repo := &model.Repo{}
	return repo, s.query.GetAsProto(repo, s.store.Space(SubspaceRepos))
}

func (s *KVRepoReadOnlyOperations) ListWorkspace(branch string) ([]*model.WorkspaceEntry, error) {
	iter := s.query.RangePrefix(s.store.Space(SubspaceWorkspace), branch)
	ws := make([]*model.WorkspaceEntry, 0)
	for iter.Advance() {
		kv := iter.MustGet()
		ent := &model.WorkspaceEntry{}
		err := proto.Unmarshal(kv.Value, ent)
		if err != nil {
			return nil, errors.ErrIndexMalformed
		}
		ws = append(ws, ent)
	}
	return ws, nil
}

func (s *KVRepoReadOnlyOperations) ReadFromWorkspace(branch, path string) (*model.WorkspaceEntry, error) {
	ent := &model.WorkspaceEntry{}
	return ent, s.query.GetAsProto(ent, s.store.Space(SubspaceWorkspace), branch, path)
}

func (s *KVRepoReadOnlyOperations) ReadBranch(branch string) (*model.Branch, error) {
	b := &model.Branch{}
	return b, s.query.GetAsProto(b, s.store.Space(SubspaceBranches), branch)
}

func (s *KVRepoReadOnlyOperations) ReadObject(addr string) (*model.Object, error) {
	obj := &model.Object{}
	return obj, s.query.GetAsProto(obj, s.store.Space(SubspaceObjects), addr)
}

func (s *KVRepoReadOnlyOperations) ReadCommit(addr string) (*model.Commit, error) {
	commit := &model.Commit{}
	return commit, s.query.GetAsProto(commit, s.store.Space(SubspaceCommits), addr)
}

func (s *KVRepoReadOnlyOperations) ListTree(addr, from string, results int) ([]*model.Entry, error) {
	iter := s.query.RangePrefix(s.store.Space(SubspaceEntries), addr, from)
	entries := make([]*model.Entry, 0)
	current := 0
	for iter.Advance() {
		entryData := iter.MustGet()
		entry := &model.Entry{}
		err := proto.Unmarshal(entryData.Value, entry)
		if err != nil {
			return nil, errors.ErrIndexMalformed
		}
		entries = append(entries, entry)
		current++
		if results != -1 && current > results {
			break
		}
	}
	return entries, nil
}

func (s *KVRepoReadOnlyOperations) ReadTreeEntry(treeAddress, name string, entryType model.Entry_Type) (*model.Entry, error) {
	entry := &model.Entry{}
	return entry, s.query.GetAsProto(entry, s.store.Space(SubspaceEntries), treeAddress, int(entryType), name)
}

func (s *KVRepoOperations) WriteToWorkspacePath(branch, path string, entry *model.WorkspaceEntry) error {
	return s.query.SetProto(entry, s.store.Space(SubspaceWorkspace), branch, path)
}

func (s *KVRepoOperations) ClearWorkspace(branch string) {
	s.query.ClearPrefix(s.store.Space(SubspaceWorkspace), branch)
}

func (s *KVRepoOperations) WriteTree(address string, entries []*model.Entry) error {
	for _, entry := range entries {
		err := s.query.SetProto(entry, s.store.Space(SubspaceEntries), address, int(entry.GetType()), entry.GetName())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *KVRepoOperations) WriteObject(addr string, object *model.Object) error {
	return s.query.SetProto(object, s.store.Space(SubspaceObjects), addr)
}

func (s *KVRepoOperations) WriteCommit(addr string, commit *model.Commit) error {
	return s.query.SetProto(commit, s.store.Space(SubspaceCommits), addr)
}

func (s *KVRepoOperations) WriteBranch(name string, branch *model.Branch) error {
	return s.query.SetProto(branch, s.store.Space(SubspaceBranches), name)
}

func (s *KVRepoOperations) DeleteBranch(name string) {
	s.query.Delete(s.store.Space(SubspaceBranches), name)
}

func (s *KVRepoOperations) WriteRepo(repo *model.Repo) error {
	return s.query.SetProto(repo, s.store.Space(SubspaceRepos))
}
