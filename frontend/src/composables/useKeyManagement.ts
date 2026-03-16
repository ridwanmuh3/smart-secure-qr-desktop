import { ref } from 'vue'
import type { KeyPairInfo } from '../types'
import {
  GenerateKeyPairCmd, ListKeyPairs, DeleteKeyPair,
  SetDefaultKeyPair, ExportPublicKey, ImportPublicKey,
  ExportPrivateKey, ImportPrivateKey,
} from '../../wailsjs/go/main/App'

export function useKeyManagement() {
  const keys = ref<KeyPairInfo[]>([])
  const isLoading = ref(false)
  const error = ref('')

  async function loadKeys() {
    isLoading.value = true
    error.value = ''
    try {
      const list = await ListKeyPairs()
      keys.value = list || []
    } catch (e: any) {
      error.value = e?.message || String(e)
    } finally {
      isLoading.value = false
    }
  }

  async function generateKey(name: string) {
    error.value = ''
    try {
      await GenerateKeyPairCmd(name)
      await loadKeys()
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function deleteKey(id: string) {
    error.value = ''
    try {
      await DeleteKeyPair(id)
      await loadKeys()
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function setDefault(id: string) {
    error.value = ''
    try {
      await SetDefaultKeyPair(id)
      await loadKeys()
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function exportKey(id: string) {
    error.value = ''
    try {
      await ExportPublicKey(id)
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function exportPrivate(id: string) {
    error.value = ''
    try {
      await ExportPrivateKey(id)
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function importKey(name: string) {
    error.value = ''
    try {
      await ImportPublicKey(name)
      await loadKeys()
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function importPrivate(name: string) {
    error.value = ''
    try {
      await ImportPrivateKey(name)
      await loadKeys()
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  return {
    keys, isLoading, error,
    loadKeys, generateKey, deleteKey, setDefault,
    exportKey, exportPrivate, importKey, importPrivate,
  }
}
