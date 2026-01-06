import axios, { type AxiosInstance } from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  Character,
  CharacterDetail,
  InventoryItem,
  EquipmentLoadout,
  AuditLog,
  GMCommandRequest,
  APIResponse,
  SpellDefinition,
  CharacterSpell,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/v1';

class APIService {
  private client: AxiosInstance;
  private token: string | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Load token from localStorage
    this.token = localStorage.getItem('access_token');
    if (this.token) {
      this.setAuthToken(this.token);
    }

    // Response interceptor for handling errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          this.logout();
          window.location.href = '/';
        }
        return Promise.reject(error);
      }
    );
  }

  setAuthToken(token: string) {
    this.token = token;
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    localStorage.setItem('access_token', token);
  }

  clearAuthToken() {
    this.token = null;
    delete this.client.defaults.headers.common['Authorization'];
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
  }

  // Authentication
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<APIResponse<LoginResponse>>('/auth/login', data);
    const loginData = response.data.data;
    this.setAuthToken(loginData.tokens.access_token);
    // Store user info (not the full loginData with tokens)
    const userInfo = {
      account_id: loginData.account_id,
      username: loginData.username,
      role: loginData.role,
      display_name: loginData.display_name
    };
    localStorage.setItem('user', JSON.stringify(userInfo));
    return loginData;
  }

  logout() {
    this.clearAuthToken();
  }

  async register(data: { email: string; username: string; password: string }): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/auth/register', data);
    return response.data.data;
  }

  // Characters
  async searchCharacters(query: string): Promise<Character[]> {
    // For now, we'll use the account's characters endpoint
    // In production, you'd want a dedicated search endpoint
    const response = await this.client.get<APIResponse<Character[]>>('/characters');
    const characters = response.data.data;
    if (query) {
      return characters.filter(c =>
        c.name.toLowerCase().includes(query.toLowerCase())
      );
    }
    return characters;
  }

  async getCharacter(id: string): Promise<CharacterDetail> {
    const response = await this.client.get<APIResponse<CharacterDetail>>(`/characters/${id}`);
    return response.data.data;
  }

  async createCharacter(data: { name: string; house: string; appearance: any }): Promise<Character> {
    const response = await this.client.post<APIResponse<Character>>('/characters', data);
    return response.data.data;
  }

  async deleteCharacter(id: string): Promise<void> {
    await this.client.delete(`/characters/${id}`);
  }

  // Inventory
  async getInventory(characterId: string): Promise<{character_id: string, items: InventoryItem[], count: number}> {
    const response = await this.client.get<APIResponse<{character_id: string, items: InventoryItem[], count: number}>>(`/characters/${characterId}/inventory`);
    return response.data.data;
  }

  async getEquipment(characterId: string): Promise<EquipmentLoadout> {
    const response = await this.client.get<APIResponse<EquipmentLoadout>>(`/characters/${characterId}/equipment`);
    return response.data.data;
  }

  // GM Commands
  async executeCommand(data: GMCommandRequest): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/admin/commands/execute', data);
    return response.data.data;
  }

  async setGrade(characterId: string, grade: number): Promise<any> {
    return this.executeCommand({
      command: 'setgrade',
      target_id: characterId,
      parameters: { grade },
    });
  }

  async teleport(characterId: string, zoneId: string, x = 0, y = 0, z = 0): Promise<any> {
    return this.executeCommand({
      command: 'teleport',
      target_id: characterId,
      parameters: { zone_id: zoneId, x, y, z },
    });
  }

  async grantItem(characterId: string, itemDefId: string, quantity = 1): Promise<any> {
    return this.executeCommand({
      command: 'grantitem',
      target_id: characterId,
      parameters: { item_def_id: itemDefId, quantity },
    });
  }

  async grantSpell(characterId: string, spellId: string): Promise<any> {
    return this.executeCommand({
      command: 'grantspell',
      target_id: characterId,
      parameters: { spell_id: spellId },
    });
  }

  async removeSpell(characterId: string, spellId: string): Promise<any> {
    return this.executeCommand({
      command: 'removespell',
      target_id: characterId,
      parameters: { spell_id: spellId },
    });
  }

  // Spells
  async getAllSpells(): Promise<{spells: SpellDefinition[], count: number}> {
    const response = await this.client.get<APIResponse<{spells: SpellDefinition[], count: number}>>('/spells');
    return response.data.data;
  }

  async getCharacterSpells(characterId: string): Promise<{character_id: string, spells: CharacterSpell[], count: number}> {
    const response = await this.client.get<APIResponse<{character_id: string, spells: CharacterSpell[], count: number}>>(`/characters/${characterId}/spells`);
    return response.data.data;
  }

  // Audit Logs
  async getCharacterAuditLogs(characterId: string, limit = 50): Promise<{character_id: string, logs: AuditLog[], count: number}> {
    const response = await this.client.get<APIResponse<{character_id: string, logs: AuditLog[], count: number}>>(
      `/admin/audit/character/${characterId}?limit=${limit}`
    );
    return response.data.data;
  }

  async getMyAuditLogs(limit = 50): Promise<{admin_id: string, logs: AuditLog[], count: number}> {
    const response = await this.client.get<APIResponse<{admin_id: string, logs: AuditLog[], count: number}>>(
      `/admin/audit/my-actions?limit=${limit}`
    );
    return response.data.data;
  }

  // Achievements
  async getAllAchievements(): Promise<{achievements: any[], count: number}> {
    const response = await this.client.get<APIResponse<{achievements: any[], count: number}>>('/achievements');
    return response.data.data;
  }

  async getCharacterAchievements(characterId: string): Promise<{achievements: any[], stats: any}> {
    const response = await this.client.get<APIResponse<{achievements: any[], stats: any}>>(`/characters/${characterId}/achievements`);
    return response.data.data;
  }

  // Friends
  async getFriends(characterId: string): Promise<{friends: any[], count: number}> {
    const response = await this.client.get<APIResponse<{friends: any[], count: number}>>(`/characters/${characterId}/friends`);
    return response.data.data;
  }

  async getPendingRequests(characterId: string): Promise<{requests: any[], count: number}> {
    const response = await this.client.get<APIResponse<{requests: any[], count: number}>>(`/characters/${characterId}/friends/requests`);
    return response.data.data;
  }

  async sendFriendRequest(requesterId: string, addresseeId: string): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/friends/request', {
      requester_id: requesterId,
      addressee_id: addresseeId,
    });
    return response.data.data;
  }

  async acceptFriendRequest(friendshipId: string): Promise<any> {
    const response = await this.client.post<APIResponse<any>>(`/friends/accept/${friendshipId}`);
    return response.data.data;
  }

  async declineFriendRequest(friendshipId: string): Promise<any> {
    const response = await this.client.post<APIResponse<any>>(`/friends/decline/${friendshipId}`);
    return response.data.data;
  }

  async removeFriend(friendshipId: string): Promise<any> {
    await this.client.delete(`/friends/${friendshipId}`);
  }

  // Leaderboards
  async getLeaderboard(limit = 50): Promise<{leaderboard: any[], count: number}> {
    const response = await this.client.get<APIResponse<{leaderboard: any[], count: number}>>(`/leaderboard?limit=${limit}`);
    return response.data.data;
  }

  async getHouseLeaderboard(house: string, limit = 50): Promise<{leaderboard: any[], count: number}> {
    const response = await this.client.get<APIResponse<{leaderboard: any[], count: number}>>(`/leaderboard/house/${house}?limit=${limit}`);
    return response.data.data;
  }

  // Account Management
  async getAllAccounts(): Promise<{accounts: any[], count: number}> {
    const response = await this.client.get<APIResponse<{accounts: any[], count: number}>>('/admin/accounts');
    return response.data.data;
  }

  async getAccountById(accountId: string): Promise<{account: any, characters: any[]}> {
    const response = await this.client.get<APIResponse<{account: any, characters: any[]}>>(`/admin/accounts/${accountId}`);
    return response.data.data;
  }

  async updateAccountStatus(accountId: string, status: string, reason: string): Promise<any> {
    const response = await this.client.put<APIResponse<any>>(`/admin/accounts/${accountId}/status`, {
      status,
      reason,
    });
    return response.data.data;
  }

  async updateAccountDiscord(accountId: string, discordId: string): Promise<any> {
    const response = await this.client.put<APIResponse<any>>(`/admin/accounts/${accountId}/discord`, {
      discord_id: discordId,
    });
    return response.data.data;
  }

  async searchAccounts(query: string): Promise<{accounts: any[], count: number}> {
    const response = await this.client.get<APIResponse<{accounts: any[], count: number}>>(`/admin/accounts/search?q=${encodeURIComponent(query)}`);
    return response.data.data;
  }

  // Notebooks
  async getCharacterNotebooks(characterId: string): Promise<{notebooks: any[]; count: number}> {
    const response = await this.client.get<APIResponse<{notebooks: any[]; count: number}>>(`/characters/${characterId}/notebooks`);
    return response.data.data;
  }

  async getNotebook(notebookId: string): Promise<any> {
    const response = await this.client.get<APIResponse<any>>(`/notebooks/${notebookId}`);
    return response.data.data;
  }

  async createNotebook(data: {character_id: string, title: string, content: string, subject?: string}): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/notebooks', data);
    return response.data.data;
  }

  async updateNotebook(notebookId: string, data: {title?: string, content?: string, subject?: string}): Promise<any> {
    const response = await this.client.put<APIResponse<any>>(`/notebooks/${notebookId}`, data);
    return response.data.data;
  }

  async deleteNotebook(notebookId: string): Promise<void> {
    await this.client.delete(`/notebooks/${notebookId}`);
  }

  // Academic Grades
  async getCharacterGrades(characterId: string, year?: number): Promise<{grades: any[]; count: number}> {
    const url = year
      ? `/characters/${characterId}/grades?year=${year}`
      : `/characters/${characterId}/grades`;
    const response = await this.client.get<APIResponse<{grades: any[]; count: number}>>(url);
    return response.data.data;
  }

  async createGrade(data: {
    character_id: string,
    school_year: number,
    subject: string,
    grade_value: string,
    exam_type: string,
    score?: number,
    teacher_comment?: string
  }): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/grades', data);
    return response.data.data;
  }

  async deleteGrade(gradeId: string): Promise<void> {
    await this.client.delete(`/grades/${gradeId}`);
  }

  // Report Cards
  async getAllReportCards(characterId: string): Promise<{report_cards: any[]; count: number}> {
    const response = await this.client.get<APIResponse<{report_cards: any[]; count: number}>>(`/characters/${characterId}/report-cards`);
    return response.data.data;
  }

  async getReportCard(characterId: string, year: number): Promise<{report_card: any}> {
    const response = await this.client.get<APIResponse<{report_card: any}>>(`/characters/${characterId}/report-cards/${year}`);
    return response.data.data;
  }

  async createReportCard(data: {
    character_id: string,
    school_year: number,
    overall_comment?: string,
    headmaster_signature?: string
  }): Promise<any> {
    const response = await this.client.post<APIResponse<any>>('/report-cards', data);
    return response.data.data;
  }

  async deleteReportCard(reportCardId: string): Promise<void> {
    await this.client.delete(`/report-cards/${reportCardId}`);
  }

  isAuthenticated(): boolean {
    return !!this.token;
  }

  getCurrentUser() {
    const userStr = localStorage.getItem('user');
    if (!userStr || userStr === 'undefined' || userStr === 'null') {
      return null;
    }
    try {
      return JSON.parse(userStr);
    } catch (e) {
      console.error('Failed to parse user from localStorage:', e);
      return null;
    }
  }
}

export const apiService = new APIService();
