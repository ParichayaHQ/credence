import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNotification } from '../contexts/NotificationContext';

// Query keys
export const queryKeys = {
  keys: ['keys'] as const,
  key: (id: string) => ['keys', id] as const,
  dids: ['dids'] as const,
  did: (id: string) => ['dids', id] as const,
  credentials: (filter?: any) => ['credentials', filter] as const,
  credential: (id: string) => ['credentials', id] as const,
  events: (filter?: any) => ['events', filter] as const,
  event: (id: string) => ['events', id] as const,
  trustScores: (did?: string) => ['trustScores', did] as const,
  networkStatus: ['networkStatus'] as const,
};

// Key Management Hooks

export function useKeys() {
  return useQuery({
    queryKey: queryKeys.keys,
    queryFn: async () => {
      const response = await window.walletAPI.keys.list();
      return response.data || [];
    },
    enabled: !!window.walletAPI,
  });
}

export function useKey(keyId: string) {
  return useQuery({
    queryKey: queryKeys.key(keyId),
    queryFn: async () => {
      const response = await window.walletAPI.keys.get(keyId);
      return response.data;
    },
    enabled: !!keyId && !!window.walletAPI,
  });
}

export function useGenerateKey() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (keyType: string) => {
      const response = await window.walletAPI.keys.generate(keyType);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.keys });
      showNotification('Key generated successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to generate key', 'error', error.message);
    },
  });
}

export function useDeleteKey() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (keyId: string) => {
      await window.walletAPI.keys.delete(keyId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.keys });
      showNotification('Key deleted successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to delete key', 'error', error.message);
    },
  });
}

// DID Management Hooks

export function useDIDs() {
  return useQuery({
    queryKey: queryKeys.dids,
    queryFn: async () => {
      const response = await window.walletAPI.dids.list();
      return response.data || [];
    },
    enabled: !!window.walletAPI,
  });
}

export function useDID(did: string) {
  return useQuery({
    queryKey: queryKeys.did(did),
    queryFn: async () => {
      const response = await window.walletAPI.dids.get(did);
      return response.data;
    },
    enabled: !!did && !!window.walletAPI,
  });
}

export function useCreateDID() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async ({ keyId, method }: { keyId: string; method: string }) => {
      const response = await window.walletAPI.dids.create(keyId, method);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.dids });
      showNotification('DID created successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to create DID', 'error', error.message);
    },
  });
}

export function useResolveDID() {
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (did: string) => {
      const response = await window.walletAPI.dids.resolve(did);
      return response.data;
    },
    onError: (error: any) => {
      showNotification('Failed to resolve DID', 'error', error.message);
    },
  });
}

// Credential Management Hooks

export function useCredentials(filter?: any) {
  return useQuery({
    queryKey: queryKeys.credentials(filter),
    queryFn: async () => {
      const response = await window.walletAPI.credentials.list(filter);
      return response.data || [];
    },
    enabled: !!window.walletAPI,
  });
}

export function useCredential(credentialId: string) {
  return useQuery({
    queryKey: queryKeys.credential(credentialId),
    queryFn: async () => {
      const response = await window.walletAPI.credentials.get(credentialId);
      return response.data;
    },
    enabled: !!credentialId && !!window.walletAPI,
  });
}

export function useStoreCredential() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async ({ credential, metadata }: { credential: any; metadata?: any }) => {
      const response = await window.walletAPI.credentials.store(credential, metadata);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.credentials() });
      showNotification('Credential stored successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to store credential', 'error', error.message);
    },
  });
}

export function useDeleteCredential() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (credentialId: string) => {
      await window.walletAPI.credentials.delete(credentialId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.credentials() });
      showNotification('Credential deleted successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to delete credential', 'error', error.message);
    },
  });
}

// Event Management Hooks

export function useEvents(filter?: any) {
  return useQuery({
    queryKey: queryKeys.events(filter),
    queryFn: async () => {
      const response = await window.walletAPI.events.list(filter);
      return response.data || [];
    },
    enabled: !!window.walletAPI,
  });
}

export function useCreateEvent() {
  const queryClient = useQueryClient();
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async ({ 
      type, 
      from, 
      to, 
      context, 
      payloadCID 
    }: { 
      type: string; 
      from: string; 
      to: string; 
      context: string; 
      payloadCID?: string; 
    }) => {
      const response = await window.walletAPI.events.create(type, from, to, context, payloadCID);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.events() });
      showNotification('Event created successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to create event', 'error', error.message);
    },
  });
}

// Trust Score Hooks

export function useTrustScores(did?: string) {
  return useQuery({
    queryKey: queryKeys.trustScores(did),
    queryFn: async () => {
      const response = await window.walletAPI.trustScores.list(did);
      return response.data || [];
    },
    enabled: !!window.walletAPI,
  });
}

export function useTrustScore(did: string, context?: string) {
  return useQuery({
    queryKey: ['trustScore', did, context],
    queryFn: async () => {
      const response = await window.walletAPI.trustScores.get(did, context);
      return response.data;
    },
    enabled: !!did && !!window.walletAPI,
  });
}

// Network Status Hook

export function useNetworkStatus() {
  return useQuery({
    queryKey: queryKeys.networkStatus,
    queryFn: async () => {
      const response = await window.walletAPI.system.getNetworkStatus();
      return response.data;
    },
    enabled: !!window.walletAPI,
    refetchInterval: 30000, // Refresh every 30 seconds
  });
}

// Network Health Monitoring Hooks

export function useNetworkHealth() {
  return useQuery({
    queryKey: ['networkHealth'],
    queryFn: async () => {
      const response = await window.walletAPI.network.getHealth();
      return response.data;
    },
    enabled: !!window.walletAPI,
    refetchInterval: 10000, // Refresh every 10 seconds
  });
}

export function useNetworkPeers() {
  return useQuery({
    queryKey: ['networkPeers'],
    queryFn: async () => {
      const response = await window.walletAPI.network.getPeers();
      return response.data || [];
    },
    enabled: !!window.walletAPI,
    refetchInterval: 15000, // Refresh every 15 seconds
  });
}

export function useNetworkStats() {
  return useQuery({
    queryKey: ['networkStats'],
    queryFn: async () => {
      const response = await window.walletAPI.network.getStats();
      return response.data;
    },
    enabled: !!window.walletAPI,
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}

// Checkpoint Hooks

export function useCheckpoints() {
  return useQuery({
    queryKey: ['checkpoints'],
    queryFn: async () => {
      const response = await window.walletAPI.checkpoints.list();
      return response.data || [];
    },
    enabled: !!window.walletAPI,
    refetchInterval: 60000, // Refresh every minute
  });
}

export function useLatestCheckpoint() {
  return useQuery({
    queryKey: ['latestCheckpoint'],
    queryFn: async () => {
      const response = await window.walletAPI.checkpoints.latest();
      return response.data;
    },
    enabled: !!window.walletAPI,
    refetchInterval: 30000, // Refresh every 30 seconds
  });
}

export function useVerifyCheckpoint() {
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (epoch: string) => {
      const response = await window.walletAPI.checkpoints.verify(epoch);
      return response.data;
    },
    onSuccess: (result) => {
      if (result.verified) {
        showNotification('Checkpoint verified successfully', 'success');
      } else {
        showNotification('Checkpoint verification failed', 'error');
      }
    },
    onError: (error: any) => {
      showNotification('Failed to verify checkpoint', 'error', error.message);
    },
  });
}

// Real-time Event Hooks

export function useEventStream() {
  return useQuery({
    queryKey: ['eventStream'],
    queryFn: async () => {
      // This would establish an SSE or WebSocket connection
      // For now, return recent events
      const response = await window.walletAPI.events.getRecent(50);
      return response.data || [];
    },
    enabled: !!window.walletAPI,
    refetchInterval: 5000, // Refresh every 5 seconds for real-time feel
  });
}

// Rules Registry Hooks

export function useActiveRules() {
  return useQuery({
    queryKey: ['activeRules'],
    queryFn: async () => {
      const response = await window.walletAPI.rules.getActive();
      return response.data;
    },
    enabled: !!window.walletAPI,
    refetchInterval: 300000, // Refresh every 5 minutes
  });
}

export function useRulesUpdates() {
  return useQuery({
    queryKey: ['rulesUpdates'],
    queryFn: async () => {
      const response = await window.walletAPI.rules.getUpdates();
      return response.data || [];
    },
    enabled: !!window.walletAPI,
    refetchInterval: 60000, // Refresh every minute
  });
}

// Wallet Operations Hooks

export function useLockWallet() {
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (password: string) => {
      await window.walletAPI.system.lock(password);
    },
    onSuccess: () => {
      showNotification('Wallet locked successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to lock wallet', 'error', error.message);
    },
  });
}

export function useUnlockWallet() {
  const { showNotification } = useNotification();

  return useMutation({
    mutationFn: async (password: string) => {
      await window.walletAPI.system.unlock(password);
    },
    onSuccess: () => {
      showNotification('Wallet unlocked successfully', 'success');
    },
    onError: (error: any) => {
      showNotification('Failed to unlock wallet', 'error', error.message);
    },
  });
}