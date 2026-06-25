'use client'

import type { FC } from 'react'
import type { VideoCodec } from 'livekit-client'
import type { LocalUserChoices } from '@livekit/components-react'
import type { ConnectionDetails } from '@/feat/types'
import type { LocalUserChoicesPassword } from '@/feat/Room'
import { useEffect, useRef, useState } from 'react'
import { RoomContent, RoomConference, InterceptorRoom, PreJoin } from '@/feat/Room'
import { ConnectionInterceptor } from '@/feat/enum'
import { useSession } from 'next-auth/react'
import { qstring } from '@/lib/utils'
import { prejoinVerify } from '@/feat/api'

const LIVEKIT_CSS_ENABLE = true

const LIVEKIT_CSS_ID = 'livekit-style'

const LIVEKIT_CSS_PATH = '/lib/css/livekit.css'

export interface RoomDetailProps {
  roomName: string
  region?: string
  hq: boolean
  codec: VideoCodec
  singlePeerConnection: boolean
  isTesting?: boolean
}

export const RoomDetail: FC<RoomDetailProps> = (props) => {
  const [interceptor, setInterceptor] = useState<ConnectionInterceptor | null>(null)
  const [isCSSLoaded, setIsCSSLoaded] = useState(!LIVEKIT_CSS_ENABLE)
  const [loading, setLoading] = useState(false)
  const [preJoinChoices, setPreJoinChoices] = useState<LocalUserChoices | undefined>()
  const [connectionDetails, setConnectionDetails] = useState<ConnectionDetails | undefined>()
  const { data: session } = useSession()
  const username = session?.profile.username ?? 'Unknown'

  // Reference
  const preJoinDefaults = useRef({ username: '', audioEnabled: false, videoEnabled: false })
  const userChoiceRef = useRef<LocalUserChoicesPassword | null>(null)
  const isReady = !!connectionDetails && !!preJoinChoices
  const handlePreJoinError = useRef((e: unknown) => console.log('Failed to handle prejoin:', e))
  const handlePreJoinSubmit = useRef(async ({ password, ...values }: LocalUserChoicesPassword) => {
    userChoiceRef.current = { password, ...values, username }

    setPreJoinChoices(values)
    setLoading(true)

    try {
      const { data: connectionDetailsData, interceptor } = await prejoinVerify({
        roomName: props.roomName,
        participantName: username,
        password,
        region: props.region,
      })

      if (connectionDetailsData) {
        setConnectionDetails(connectionDetailsData)
      } else {
        setInterceptor(interceptor)
      }
    } catch (e) {
      setInterceptor(ConnectionInterceptor.Unknown)
      console.log('Failed to join the room:', e)
    } finally {
      setLoading(false)
    }
  })

  const handleBackToPrejoin = useRef(() => {
    setInterceptor(null)
    setPreJoinChoices(undefined)
  })

  useEffect(() => {
    if (!isReady) {
      const livekitStyle = document.getElementById(LIVEKIT_CSS_ID)

      if (!livekitStyle) {
        return
      }

      document.head.removeChild(livekitStyle)
    } else {
      if (LIVEKIT_CSS_ENABLE) {
        const livekitCss = document.createElement('link')

        livekitCss.rel = 'stylesheet'
        livekitCss.id = LIVEKIT_CSS_ID
        livekitCss.href = window.location.origin + LIVEKIT_CSS_PATH
        livekitCss.onload = () => {
          setInterceptor(null) // Ready to live after css is fully loaded
          setIsCSSLoaded(true)
        }

        document.head.appendChild(livekitCss)
      }
    }
  }, [isReady])

  useEffect(() => {
    if (interceptor === ConnectionInterceptor.Pending) {
      const url = (process.env.NEXT_PUBLIC_BACKEND_URL ?? '') + '/api/waiting-rooms/request'
      const es = new EventSource(
        qstring(
          url,
          { room_code: props.roomName, token: session?.access_token },
          { skipEmpty: true }
        )
      )

      es.onmessage = async (e: MessageEvent<string>) => {
        const { status }: { status: string } = JSON.parse(e.data)

        if (!userChoiceRef.current) return
        if (status === 'accepted') {
          await handlePreJoinSubmit.current(userChoiceRef.current)
          es.close()
        }
        if (status === 'rejected') {
          // @TODO
          alert('Kamu telah di tolak untuk join ruangan')

          handleBackToPrejoin.current()
          es.close()
        }
      }

      es.onerror = (e) => {
        console.log(e)
        // @TODO
        // alert('Tidak dapat mengakses ruangan saat ini')

        // handleBackToPrejoin.current()
        // es.close()
      }

      return () => es.close()
    }
  }, [interceptor])

  if (interceptor) {
    return <InterceptorRoom interceptor={interceptor} onClick={handleBackToPrejoin.current} />
  }

  return isReady && isCSSLoaded ? (
    <RoomConference
      connectionDetails={connectionDetails}
      userChoices={preJoinChoices}
      options={{
        codec: props.codec,
        hq: props.hq,
        singlePeerConnection: props.singlePeerConnection,
      }}
      roomCode={props.roomName}
      appToken={session?.access_token ?? ''}
    >
      <RoomContent />
    </RoomConference>
  ) : (
    <PreJoin
      defaults={{ ...preJoinDefaults.current, username }}
      onSubmit={handlePreJoinSubmit.current}
      onError={handlePreJoinError.current}
      isLoading={loading}
      isGuest={false}
    />
  )
}
