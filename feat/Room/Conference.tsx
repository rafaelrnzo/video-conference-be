'use client'

import type { FC, ReactNode } from 'react'
import type {
  RoomOptions,
  TrackPublishDefaults,
  VideoCaptureOptions,
  VideoCodec,
} from 'livekit-client'
import type { LocalUserChoices } from '@livekit/components-react'
import type { ConnectionDetails } from '@/feat/types'
import { useEffect, useMemo, useRef } from 'react'
import { ConnectionState, MediaDeviceFailure, Room, RoomEvent, VideoPresets } from 'livekit-client'
import { RoomContext } from '@livekit/components-react'
import { useParamsState } from '@/hooks'
import { RoomState, RoomLayout } from '@/feat/Room'

export interface RoomConferenceProps {
  children?: ReactNode
  userChoices: LocalUserChoices
  connectionDetails: ConnectionDetails
  options: {
    hq: boolean
    codec: VideoCodec
    singlePeerConnection: boolean
  }
  roomCode: string
  appToken: string
}

export const RoomConference: FC<RoomConferenceProps> = ({ children, ...props }) => {
  const propsRef = useRef(props)
  const roomOptions = useRef((): RoomOptions => {
    const { current } = propsRef
    const videoCodec: VideoCodec | undefined = current.options.codec ?? 'vp9'
    const videoCaptureDefaults: VideoCaptureOptions = {
      deviceId: current.userChoices.videoDeviceId ?? undefined,
      resolution: current.options.hq ? VideoPresets.h2160 : VideoPresets.h720,
    }
    const publishDefaults: TrackPublishDefaults = {
      dtx: false,
      videoSimulcastLayers: current.options.hq
        ? [VideoPresets.h1080, VideoPresets.h720]
        : [VideoPresets.h540, VideoPresets.h216],
      red: true,
      videoCodec,
    }

    return {
      videoCaptureDefaults: videoCaptureDefaults,
      publishDefaults: publishDefaults,
      audioCaptureDefaults: {
        deviceId: current.userChoices.audioDeviceId ?? undefined,
      },
      adaptiveStream: true,
      dynacast: true,
      singlePeerConnection: current.options.singlePeerConnection,
    }
  })

  const { router } = useParamsState<{ name: string }>()
  const room = useMemo(() => new Room(roomOptions.current()), []) // Maybe changed
  const roomEvent = useRef({
    leave: () => router.replace('/'),
    error: (e: unknown) => {
      console.log('Failed to get active media devices:', e)

      alert(
        `Encountered an unexpected error, check the console logs for details: ${(e as Error).message}`
      )
    },
  })

  useEffect(() => {
    const { serverUrl, participantToken } = propsRef.current.connectionDetails
    const { leave, error } = roomEvent.current

    room.on(RoomEvent.Disconnected, leave)
    room.on(RoomEvent.MediaDevicesError, error)
    room.on(RoomEvent.MediaDevicesError, (error) => {
      const failure = MediaDeviceFailure.getFailure(error)

      if (failure === MediaDeviceFailure.PermissionDenied) {
        console.log('User disallowed access to the capturing device.')
      } else if (failure === MediaDeviceFailure.NotFound) {
        console.log('The requested device is unavailable.')
      }
    })

    let mounted = true

    async function connect() {
      try {
        await room.connect(serverUrl, participantToken, {
          autoSubscribe: true,
        })

        if (!mounted) return

        if (propsRef.current.userChoices.videoEnabled) {
          await room.localParticipant.setCameraEnabled(true)
        }

        if (propsRef.current.userChoices.audioEnabled) {
          await room.localParticipant.setMicrophoneEnabled(true)
        }
      } catch (e) {
        console.log('Failed to connect to the room:', e)
      }
    }

    connect()

    return () => {
      mounted = false

      room.off(RoomEvent.Disconnected, leave)
      room.off(RoomEvent.MediaDevicesError, error)

      if (
        room.state === ConnectionState.Connected ||
        room.state === ConnectionState.Connecting ||
        room.state === ConnectionState.Reconnecting
      ) {
        if (propsRef.current.appToken && propsRef.current.roomCode) {
          void fetch((process.env.NEXT_PUBLIC_BACKEND_URL || '') + '/api/livekit/leave', {
            method: 'POST',
            keepalive: true,
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${propsRef.current.appToken}`,
            },
            body: JSON.stringify({ room_code: propsRef.current.roomCode }),
          })
        }
        room.disconnect()
      }
    }
  }, [room])

  return (
    <RoomContext.Provider value={room}>
      <RoomState>
        <RoomLayout>{children}</RoomLayout>
      </RoomState>
    </RoomContext.Provider>
  )
}
