export type FileStatus = 'pending' | 'ready' | 'quarantined' | 'deleted';

export type File = {
  id: string;
  ownerId: string;
  name: string;
  size: number;
  mimeType: string;
  status: FileStatus;
  storageKey: string;
  createdAt: string;
  updatedAt: string;
};
