export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  account: Account;
  characters: Character[];
}

export interface Account {
  id: string;
  email: string;
  username: string;
  role: string;
  created_at: string;
}

export interface Character {
  id: string;
  account_id: string;
  name: string;
  house: string;
  grade: number;
  zone_id: string;
  level: number;
  experience: number;
  created_at: string;
}

export interface CharacterDetail extends Character {
  stats: CharacterStats;
  currencies: Record<string, number>;
  permissions: string[];
}

export interface CharacterStats {
  health_current: number;
  health_max: number;
  mana_current: number;
  mana_max: number;
  stamina_current: number;
  stamina_max: number;
  strength: number;
  dexterity: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
  luck: number;
  spell_power: number;
  defense: number;
  speed: number;
}

export interface InventoryItem {
  id: string;
  item_def_id: string;
  quantity: number;
  display_name: string;
  description: string;
  item_type: string;
  equipment_slot?: string;
  is_stackable: boolean;
  max_stack_size: number;
  base_value: number;
}

export interface EquipmentLoadout {
  id: string;
  character_id: string;
  loadout_name: string;
  is_active: boolean;
  slot_wand?: string;
  slot_robe?: string;
  slot_hat?: string;
  slot_cloak?: string;
  slot_amulet?: string;
  slot_ring_left?: string;
  slot_ring_right?: string;
  slot_boots?: string;
  slot_gloves?: string;
  slot_broom?: string;
}

export interface AuditLog {
  id: string;
  event_type: string;
  actor_id: string;
  actor_type: string;
  target_id: string;
  target_type: string;
  action: string;
  details: string;
  old_value?: string;
  new_value?: string;
  ip_address?: string;
  success: boolean;
  error_message?: string;
  created_at: string;
}

export interface GMCommandRequest {
  command: string;
  target_id: string;
  parameters: Record<string, any>;
}

export interface APIResponse<T> {
  success: boolean;
  data: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
}
