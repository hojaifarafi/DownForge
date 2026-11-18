import { useState, useEffect, useRef, useCallback } from 'react';

export function useWebSocket(onEvent) {
  const [isConnected, setIsConnected] = useState(false);
  const [downloads, setDownloads] = useState([]);
  const [stats, setStats] = useState({
    total_downloads: 0,
    active_downloads: 0,
    completed_downloads: 0,
    paused_downloads: 0,
    failed_downloads: 0,
    total_download_speed: 0,
    total_upload_speed: 0,
  });

  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const connect = useCallback(() => {
    // Clean up any existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use window.location.host or fallback for dev proxy
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Auto reconnect after 2 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 2000);
    };

    ws.onerror = () => {
      ws.close();
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.event === 'progress' && Array.isArray(msg.payload)) {
          setDownloads(msg.payload);
        } else if (msg.event === 'stats' && msg.payload) {
          setStats(msg.payload);
        }

        if (onEvent) {
          onEvent(msg.event, msg.payload);
        }
      } catch (err) {
        console.error('Error parsing WebSocket message:', err);
      }
    };
  }, [onEvent]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    isConnected,
    downloads,
    setDownloads,
    stats,
    setStats,
  };
}
