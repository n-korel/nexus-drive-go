import { useEffect, useState } from "react";
import { WEBSOCKET_URL } from "../constants";
import { Trip, Driver, CarPackageSlug } from "../types";
import {
  TripEvents,
  isValidWsMessage,
  isValidTripEvent,
  ClientWsMessage,
  BackendEndpoints,
} from "../contracts";

interface useDriverConnectionProps {
  location: {
    latitude: number;
    longitude: number;
  };
  geohash: string;
  userID: string;
  packageSlug: CarPackageSlug;
}

export const useDriverStreamConnection = ({
  location,
  geohash,
  userID,
  packageSlug,
}: useDriverConnectionProps) => {
  const [requestedTrip, setRequestedTrip] = useState<Trip | null>(null);
  const [tripStatus, setTripStatus] = useState<TripEvents | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [driver, setDriver] = useState<Driver | null>(null);

  useEffect(() => {
    if (!userID) return;

    const websocket = new WebSocket(
      `${WEBSOCKET_URL}${BackendEndpoints.WS_DRIVERS}?userID=${userID}&packageSlug=${packageSlug}`
    );
    setWs(websocket);

    websocket.onopen = () => {
      if (location) {
        websocket.send(
          JSON.stringify({
            type: TripEvents.DriverLocation,
            data: {
              location,
              geohash,
            },
          })
        );
      }
    };

    websocket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);

        if (!isValidWsMessage(message)) {
          setError(`Unknown message type "${message?.type}"`);
          return;
        }

        switch (message.type) {
          case TripEvents.DriverTripRequest:
            const tripData = message.data as Trip;
            setRequestedTrip(tripData);
            setTripStatus(message.type);
            break;
          case TripEvents.DriverRegister:
            setDriver(message.data as Driver);
            setTripStatus(message.type);
            break;
          default:
            if (isValidTripEvent(message.type)) {
              setTripStatus(message.type);
            }
        }
      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
        setError("Failed to parse server message");
      }
    };

    websocket.onclose = () => {
      console.log("WebSocket closed");
      setWs(null);
    };

    websocket.onerror = (event) => {
      setError("WebSocket error occurred");
      console.error("WebSocket error:", event);
    };

    return () => {
      console.log("Closing WebSocket");
      if (websocket.readyState === WebSocket.OPEN) {
        websocket.close();
      }
    };
  }, [userID, packageSlug, location, geohash]);

  const sendMessage = (message: ClientWsMessage) => {
    if (ws?.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify(message));
      } catch (err) {
        console.error("Error sending message:", err);
        setError("Failed to send message");
      }
    } else {
      setError("WebSocket is not connected");
    }
  };

  const resetTripStatus = () => {
    setTripStatus(null);
    setRequestedTrip(null);
  };

  return {
    error,
    tripStatus,
    driver,
    requestedTrip,
    resetTripStatus,
    sendMessage,
    setTripStatus,
  };
};
