import api from './api';
import {
  InstantQueryParams,
  RangeQueryParams,
  QueryResponse,
  RangeQueryResponse,
  QueryValidation
} from '../types/queries';

/**
 * Service for interacting with the queries API
 */
export const queriesService = {
  /**
   * Execute an instant query
   * @param params Query parameters
   */
  async executeInstantQuery(params: InstantQueryParams): Promise<QueryResponse> {
    const response = await api.post('/query', params);
    return response.data;
  },
  
  /**
   * Execute a range query
   * @param params Query parameters
   */
  async executeRangeQuery(params: RangeQueryParams): Promise<RangeQueryResponse> {
    const response = await api.post('/query/range', params);
    return response.data;
  },
  
  /**
   * Validate a query without executing it
   * @param query The query to validate
   */
  async validateQuery(query: string): Promise<QueryValidation> {
    const response = await api.post('/query/validate', { query });
    return response.data;
  },
  
  /**
   * Get query suggestions
   * @param prefix Optional prefix to filter suggestions
   * @param limit Maximum number of suggestions
   */
  async getQuerySuggestions(prefix: string = '', limit: number = 10): Promise<string[]> {
    const response = await api.get('/query/suggestions', {
      params: { prefix, limit }
    });
    return response.data.suggestions;
  },
  
  /**
   * Execute an instant query with error handling
   * @param query The PromQL query
   * @param time Optional timestamp (ISO string)
   */
  async executeQuery(query: string, time?: string): Promise<QueryResponse> {
    try {
      return await this.executeInstantQuery({ query, time });
    } catch (error) {
      console.error('Query execution error:', error);
      throw new Error(`Failed to execute query: ${query}`);
    }
  },
  
  /**
   * Execute a range query with sensible defaults
   * @param query The PromQL query
   * @param start Start time
   * @param end End time
   * @param step Step in seconds (default: 60)
   */
  async executeTimeSeriesQuery(
    query: string,
    start: Date,
    end: Date,
    step: number = 60
  ): Promise<RangeQueryResponse> {
    try {
      return await this.executeRangeQuery({
        query,
        start: start.toISOString(),
        end: end.toISOString(),
        step
      });
    } catch (error) {
      console.error('Time series query execution error:', error);
      throw new Error(`Failed to execute time series query: ${query}`);
    }
  },
  
  /**
   * Get a simple value from a query (useful for gauges)
   * @param query The PromQL query
   */
  async getScalarValue(query: string): Promise<number | null> {
    try {
      const result = await this.executeInstantQuery({ query });
      if (result.data && result.data.length > 0) {
        return result.data[0].value;
      }
      return null;
    } catch (error) {
      console.error('Error getting scalar value:', error);
      return null;
    }
  }
};

export default queriesService;