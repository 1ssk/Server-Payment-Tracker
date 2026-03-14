export interface User {
  username: string;
}

export interface SMTPSettings {
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  to: string;
  enabled: boolean;
}
