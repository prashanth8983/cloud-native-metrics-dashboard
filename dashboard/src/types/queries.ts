export interface InstantQueryParams {
    query: string;
    time?: string;
  }
  
  export interface RangeQueryParams {
    query: string;
    start: string;
    end: string;
    step: number;
  }
  
  export interface QueryResponse {
    query: string;
    queryTime: string;
    status: string;
    data: DataPoint[];
  }
  
  export interface DataPoint {
    metricName: string;
    labels: Record<string, string>;
    value: number;
    timestamp: string;
  }
  
  export interface RangeQueryResponse {
    query: string;
    start: string;
    end: string;
    step: number;
    status: string;
    series: TimeSeries[];
  }
  
  export interface TimeSeries {
    metricName: string;
    labels: Record<string, string>;
    dataPoints: TimeValuePair[];
  }
  
  export interface TimeValuePair {
    timestamp: string;
    value: number;
  }
  
  export interface QueryValidation {
    query: string;
    valid: boolean;
    message: string;
  }