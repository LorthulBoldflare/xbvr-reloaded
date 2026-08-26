import type { File } from '../api/types'

// Derive a search query / scene title from a video filename — port of the
// old SceneMatch/CreateScene cleanup.
const COMMON_WORDS = new Set([
  '180', '180x180', '2880x1440', '3d', '3dh', '3dv', '30fps', '30m', '360',
  '3840x1920', '4k', '5k', '5400x2700', '60fps', '6k', '7k', '7680x3840',
  '8k', 'fb360', 'fisheye190', 'funscript', 'cmscript', 'h264', 'h265', 'hevc',
  'hq', 'hsp', 'lq', 'lr', 'mkv', 'mkx200', 'mkx220', 'mono', 'mp4', 'oculus',
  'oculus5k', 'oculusrift', 'original', 'rf52', 'smartphone', 'srt', 'ssa',
  'tb', 'uhq', 'vrca220', 'vp9'
])

export function cleanFilename(file: File): string {
  return file.filename
    .replace(/[._+'’`-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .split(' ')
    .filter((w) => !COMMON_WORDS.has(w.toLowerCase()) && !/^[0-9]+p$/.test(w))
    .join(' ')
}

export function cleanFilenameForTitle(file: File): string {
  return cleanFilename(file).replace(/ s /g, "'s ")
}
