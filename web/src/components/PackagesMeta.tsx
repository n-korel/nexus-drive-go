import { Bus, Truck, Crown } from "lucide-react";
import { Car } from "lucide-react";
import { CarPackageSlug } from "../types";

export const PackagesMeta: Record<
  CarPackageSlug,
  {
    name: string;
    icon: React.ReactNode;
    description: string;
  }
> = {
  [CarPackageSlug.SEDAN]: {
    name: "Седан",
    icon: <Car />,
    description: "Экономичный и комфортный вариант",
  },
  [CarPackageSlug.SUV]: {
    name: "Внедорожник",
    icon: <Truck />,
    description: "Просторный вариант для компаний",
  },
  [CarPackageSlug.VAN]: {
    name: "Минивэн",
    icon: <Bus />,
    description: "Идеален для больших групп",
  },
  [CarPackageSlug.LUXURY]: {
    name: "Премиум",
    icon: <Crown />,
    description: "Максимальный комфорт и стиль",
  },
};
