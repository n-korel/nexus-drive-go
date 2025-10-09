import { Trip } from "../types";
import { TripOverviewCard } from "./TripOverviewCard";
import { Button } from "./ui/button";
import { TripEvents } from "../contracts";

interface DriverTripOverviewProps {
  trip?: Trip | null;
  status?: TripEvents | null;
  onAcceptTrip?: () => void;
  onDeclineTrip?: () => void;
}

export const DriverTripOverview = ({
  trip,
  status,
  onAcceptTrip,
  onDeclineTrip,
}: DriverTripOverviewProps) => {
  if (!trip) {
    return (
      <TripOverviewCard
        title="Ожидание пассажира..."
        description="Ожидаем, пока пассажир запросит поездку..."
      />
    );
  }

  if (status === TripEvents.DriverTripRequest) {
    return (
      <TripOverviewCard
        title="Новый запрос на поездку!"
        description="Поступил запрос на поездку. Проверьте маршрут и подтвердите, если можете принять заказ."
      >
        <div className="flex flex-col gap-2">
          <Button onClick={onAcceptTrip}>Принять поездку</Button>
          <Button variant="outline" onClick={onDeclineTrip}>
            Отклонить поездку
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.DriverTripAccept) {
    return (
      <TripOverviewCard
        title="Готово!"
        description="Теперь вы можете начать поездку."
      >
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <h3 className="text-lg font-bold">Детали поездки</h3>
            <p className="text-sm text-gray-500">
              ID поездки: {trip.id}
              <br />
              ID пассажира: {trip.userID}
            </p>
          </div>
        </div>
      </TripOverviewCard>
    );
  }

  return null;
};
