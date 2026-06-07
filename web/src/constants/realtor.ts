export interface BuildingType {
  id: number
  label: string
}

export const BUILDING_TYPES: BuildingType[] = [
  { id: 1, label: 'House' },
  { id: 2, label: 'Duplex' },
  { id: 3, label: 'Triplex' },
  { id: 16, label: 'Row/Townhouse' },
  { id: 17, label: 'Apartment' },
  { id: 19, label: 'Fourplex' },
]

export const LISTING_FILTER_BUILDING_TYPES: BuildingType[] = [
  { id: 1, label: 'House' },
  { id: 2, label: 'Duplex' },
  { id: 3, label: 'Triplex' },
  { id: 19, label: 'Fourplex' },
]

export interface RangeOption {
  label: string
  value: string
}

export const BED_BATH_OPTIONS: RangeOption[] = [
  { label: 'Any', value: '' },
  { label: '1+', value: '1-0' },
  { label: '2+', value: '2-0' },
  { label: '3+', value: '3-0' },
  { label: '4+', value: '4-0' },
  { label: '5+', value: '5-0' },
]
