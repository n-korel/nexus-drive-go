import { RouteFare, TripPreview, Driver } from "../types";
import { DriverList } from "./DriversList";
import { Card } from "./ui/card";
import { Button } from "./ui/button";
import {
  convertMetersToKilometers,
  convertSecondsToMinutes,
} from "../utils/math";
import { Skeleton } from "./ui/skeleton";
import { TripOverviewCard } from "./TripOverviewCard";
import { StripePaymentButton } from "./StripePaymentButton";
import { DriverCard } from "./DriverCard";
import { TripEvents, PaymentEventSessionCreatedData } from "../contracts";

interface TripOverviewProps {
  trip: TripPreview | null;
  status: TripEvents | null;
  assignedDriver?: Driver | null;
  paymentSession?: PaymentEventSessionCreatedData | null;
  onPackageSelect: (carPackage: RouteFare) => void;
  onCancel: () => void;
}

export const RiderTripOverview = ({
  trip,
  status,
  assignedDriver,
  paymentSession,
  onPackageSelect,
  onCancel,
}: TripOverviewProps) => {
  if (!trip) {
    return (
      <TripOverviewCard
        title="Начните поездку"
        description="Нажмите на карту, чтобы выбрать пункт назначения"
      />
    );
  }

  if (status === TripEvents.PaymentSessionCreated && paymentSession) {
    return (
      <TripOverviewCard
        title="Требуется оплата"
        description="Пожалуйста, завершите оплату, чтобы подтвердить поездку"
      >
        <div className="flex flex-col gap-4">
          <DriverCard driver={assignedDriver} />

          <div className="text-sm text-gray-500">
            <p>
              Сумма: {paymentSession.amount} {paymentSession.currency}
            </p>
            <p>ID поездки: {paymentSession.tripID}</p>
          </div>
          <StripePaymentButton paymentSession={paymentSession} />
        </div>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.NoDriversFound) {
    return (
      <TripOverviewCard
        title="Водители не найдены"
        description="Не удалось найти водителей для вашей поездки. Пожалуйста, попробуйте позже."
      >
        <Button variant="outline" className="w-full" onClick={onCancel}>
          Назад
        </Button>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.DriverAssigned) {
    return (
      <TripOverviewCard
        title="Водитель назначен!"
        description="Ваш водитель в пути. Ожидаем подтверждения оплаты..."
      >
        <div className="flex flex-col space-y-3 justify-center items-center mb-4" />
        <Button variant="destructive" className="w-full" onClick={onCancel}>
          Отменить поездку
        </Button>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.Completed) {
    return (
      <TripOverviewCard
        title="Поездка завершена!"
        description="Спасибо, что воспользовались нашим сервисом!"
      >
        <Button variant="outline" className="w-full" onClick={onCancel}>
          Назад
        </Button>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.Cancelled) {
    return (
      <TripOverviewCard
        title="Поездка отменена"
        description="Поездка отменена. Пожалуйста, попробуйте позже."
      >
        <Button variant="outline" className="w-full" onClick={onCancel}>
          Назад
        </Button>
      </TripOverviewCard>
    );
  }

  if (status === TripEvents.Created) {
    return (
      <TripOverviewCard
        title="Поиск водителя"
        description="Ваша поездка подтверждена. Мы подбираем водителя — это займёт немного времени."
      >
        <div className="flex flex-col space-y-3 justify-center items-center mb-4">
          <Skeleton className="h-[125px] w-[250px] rounded-xl" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-[250px]" />
            <Skeleton className="h-4 w-[200px]" />
          </div>
        </div>

        <div className="flex flex-col items-center justify-center gap-2">
          {trip?.duration && (
            <h3 className="text-sm font-medium text-gray-700 mb-2">
              Прибытие через {convertSecondsToMinutes(trip?.duration)} мин,
              расстояние {convertMetersToKilometers(trip?.distance ?? 0)} км
            </h3>
          )}

          <Button variant="destructive" className="w-full" onClick={onCancel}>
            Отменить
          </Button>
        </div>
      </TripOverviewCard>
    );
  }

  if (trip.rideFares && trip.rideFares.length >= 0 && !trip.tripID) {
    return (
      <DriverList
        trip={trip}
        onPackageSelect={onPackageSelect}
        onCancel={onCancel}
      />
    );
  }

  return (
    <Card className="w-full md:max-w-[500px] z-[9999] flex-[0.3]">
      Нет данных о тарифах поездки. Обновите страницу.
    </Card>
  );
};
