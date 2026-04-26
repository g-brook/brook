export interface IpStrategy {
  id: number;
  name: string;
  type: string;
  status: number;
  created_at: string;
  updated_at: string;
  tunnels?: any[];
}

export interface IpRule {
  id: number;
  strategyId: number;
  ip: string;
  remark: string;
  created_at?: string;
}
