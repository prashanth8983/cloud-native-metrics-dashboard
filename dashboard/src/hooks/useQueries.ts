import { useState, useCallback } from 'react';
import queriesService from '../services/queriesService';
import { 
  QueryResponse, 
  RangeQueryResponse, 
  QueryValidation 
} from '../types/queries';

export function useQueries() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const executeInstantQuery = useCallback(async (query: string, time?: string): Promise<QueryResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await queriesService.executeInstantQuery({ query, time });
      return response;
    } catch (err) {
      setError(err instanceof Error ? err : new Error(`Failed to execute query: ${query}`));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const executeRangeQuery = useCallback(async (
    query: string,
    start: Date,
    end: Date,
    step: number = 60
  ): Promise<RangeQueryResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await queriesService.executeRangeQuery({
        query,
        start: start.toISOString(),
        end: end.toISOString(),
        step
      });
      return response;
    } catch (err) {
      setError(err instanceof Error ? err : new Error(`Failed to execute range query: ${query}`));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const validateQuery = useCallback(async (query: string): Promise<QueryValidation | null> => {
    try {
      return await queriesService.validateQuery(query);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(`Failed to validate query: ${query}`));
      return null;
    }
  }, []);

  const getQuerySuggestions = useCallback(async (prefix: string = '', limit: number = 10): Promise<string[]> => {
    try {
      return await queriesService.getQuerySuggestions(prefix, limit);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to get query suggestions'));
      return [];
    }
  }, []);

  return {
    loading,
    error,
    executeInstantQuery,
    executeRangeQuery,
    validateQuery,
    getQuerySuggestions
  };
}