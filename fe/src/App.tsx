import { useMemo, useState } from 'react';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import goldData from './prices.csv';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import silverData from './silver-prices.csv';
import './App.css';
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { DateTime } from 'luxon';
import { TooltipProps } from 'recharts/types/component/Tooltip';

interface PriceData {
  date: number;
  traditional: number;
  k18: number;
  k21: number;
  k22: number;
}

interface CSVPriceData {
  date: string;
  traditional: string;
  k18: string;
  k21: string;
  k22: string;
}

function formatPrice({
  price,
  isBhori = false,
  showFraction = false,
}: {
  price: number;
  isBhori?: boolean;
  showFraction?: boolean;
}) {
  return Intl.NumberFormat('en-US', {
    currency: 'BDT',
    style: 'currency',
    currencyDisplay: 'narrowSymbol',
    maximumFractionDigits: showFraction ? 1 : 0,
    notation: 'standard',
  }).format(price * (isBhori ? 11.664 : 1));
}

const filters = ['7 days', '30 days', '1 year', 'all time'] as const;
const units = ['gram', 'bhori'] as const;
const metals = ['gold', 'silver'] as const;

function App() {
  const [filter, setFilter] = useState<(typeof filters)[number]>('all time');
  const [unit, setUnit] = useState<(typeof units)[number]>('bhori');
  const [metal, setMetal] = useState<(typeof metals)[number]>('gold');

  const data = metal === 'gold' ? goldData : silverData;

  const priceData: PriceData[] = useMemo(
    () =>
      data
        .map((d: CSVPriceData) => ({
          date: DateTime.fromISO(d.date).toMillis(),
          k22: Number(d.k22),
          k21: Number(d.k21),
          k18: Number(d.k18),
          traditional: Number(d.traditional),
        }))
        .filter((d: PriceData, _: number, array: PriceData[]) => {
          const lastDate = array[array.length - 1].date;
          switch (filter) {
            case '1 year':
              return d.date > DateTime.fromMillis(lastDate).minus({ year: 1 }).toMillis();
            case '30 days':
              return d.date > DateTime.fromMillis(lastDate).minus({ days: 30 }).toMillis();
            case '7 days':
              return d.date > DateTime.fromMillis(lastDate).minus({ days: 7 }).toMillis();
            default:
              return true;
          }
        })
        .sort((a: PriceData, b: PriceData) => a.date - b.date),
    [filter, data]
  );

  const priceDataMap = useMemo(
    () =>
      priceData.reduce((acc: Record<number, PriceData>, d) => {
        acc[d.date] = d;
        return acc;
      }, {}),
    [priceData]
  );

  const isBhori = unit === 'bhori';

  function renderToolTip(data: TooltipProps<string, number>) {
    const priceData = priceDataMap[data.label];

    if (!priceData) return null;

    const { k18, k21, k22, traditional } = priceData;

    function getFormattedPrice(price: number) {
      return formatPrice({ price, isBhori, showFraction: isBhori });
    }

    return (
      <div className='tooltip'>
        <span className='tooltip-date'>
          {DateTime.fromMillis(data.label).toFormat('LLL dd, yyyy')}
        </span>{' '}
        <br />
        <span>22K: {getFormattedPrice(k22)}</span> <br />
        <span>21K: {getFormattedPrice(k21)}</span> <br />
        <span>18K: {getFormattedPrice(k18)}</span> <br />
        <span>সনাতন: {getFormattedPrice(traditional)}</span> <br />
      </div>
    );
  }

  const lastPrice = priceData[priceData.length - 1];

  const stats7Days = useMemo(() => {
    if (!lastPrice || data.length < 2) return null;

    const lastDate = DateTime.fromMillis(lastPrice.date);
    const targetDate = lastDate.minus({ days: 7 }).toMillis();

    // Find the closest data point to 7 days ago
    const dataPoints = data.map((d: CSVPriceData) => ({
      date: DateTime.fromISO(d.date).toMillis(),
      k22: Number(d.k22),
    }));

    const price7DaysAgoPoint = dataPoints
      .filter((d: { date: number }) => d.date <= targetDate)
      .sort((a: { date: number }, b: { date: number }) => b.date - a.date)[0];

    if (!price7DaysAgoPoint) return null;

    const diff = lastPrice.k22 - price7DaysAgoPoint.k22;
    const percentage = (diff / price7DaysAgoPoint.k22) * 100;

    return {
      diff,
      percentage: percentage.toFixed(2),
      isUp: diff >= 0,
    };
  }, [lastPrice, data]);

  return (
    <div className='container'>
      <h1>{metal === 'gold' ? 'Gold' : 'Silver'} Price History in Bangladesh</h1>
      {stats7Days && (
        <div className={`stats-7days ${stats7Days.isUp ? 'up' : 'down'}`}>
          <span>গত ৭ দিনে: </span>
          <strong>
            {stats7Days.isUp ? '▲' : '▼'} {formatPrice({ price: Math.abs(stats7Days.diff), isBhori })} ({stats7Days.percentage}%)
          </strong>
        </div>
      )}
      <ResponsiveContainer width='98%' aspect={2}>
        <AreaChart data={priceData}>
          <defs>
            <linearGradient id='goldGrad' x1='0' y1='0' x2='0' y2='1'>
              <stop offset='5%' stopColor='#FFD700' stopOpacity={0.8} />
              <stop offset='95%' stopColor='#ffd7004f' stopOpacity={0} />
            </linearGradient>
            <linearGradient id='silverGrad' x1='0' y1='0' x2='0' y2='1'>
              <stop offset='5%' stopColor='#C0C0C0' stopOpacity={0.8} />
              <stop offset='95%' stopColor='#c0c0c04f' stopOpacity={0} />
            </linearGradient>
          </defs>
          <Area
            type='bump'
            dataKey='k22'
            stroke={metal === 'gold' ? '#FFD700' : '#C0C0C0'}
            fillOpacity={1}
            fill={metal === 'gold' ? 'url(#goldGrad)' : 'url(#silverGrad)'}
          />
          <XAxis
            dataKey='date'
            type='number'
            domain={['dataMin', 'dataMax']}
            tickFormatter={d =>
              DateTime.fromMillis(d).toFormat(
                filter === '7 days' || filter === '30 days' ? 'LLL dd' : 'LLL yyyy'
              )
            }
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            domain={['auto']}
            tickFormatter={d => formatPrice({ price: d, isBhori, showFraction: false })}
          />
          <Tooltip content={renderToolTip} />
        </AreaChart>
      </ResponsiveContainer>
      <div className='button-group'>
        {metals.map(m => (
          <button
            className={`${m === metal ? 'active' : ''} ${m}`}
            key={m}
            onClick={() => setMetal(m)}
            style={{ textTransform: 'capitalize' }}
          >
            {m}
          </button>
        ))}
      </div>
      <div className='button-group'>
        {filters.map(f => (
          <button
            className={f === filter ? 'active' : ''}
            key={f}
            onClick={() => setFilter(f)}
          >
            {f}
          </button>
        ))}
      </div>
      <div className='button-group'>
        {units.map(u => (
          <button
            className={u === unit ? 'active' : ''}
            key={u}
            onClick={() => setUnit(u)}
          >
            {u}
          </button>
        ))}
      </div>
      <table>
        <thead>
          <tr>
            <th>22K</th>
            <th>21K</th>
            <th>18K</th>
            <th>সনাতন</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            {(['k22', 'k21', 'k18', 'traditional'] as const).map(k => (
              <td key={k}>
                {formatPrice({
                  price: lastPrice[k],
                  isBhori,
                })}
              </td>
            ))}
          </tr>
        </tbody>
      </table>
      <p className='last-price'>
        Last updated: {DateTime.fromMillis(lastPrice.date).toFormat('LLLL dd, yyyy')}{' '}
      </p>
      <p className='info'>
        * Prices are collected from{' '}
        <a
          target='_blank'
          rel='noreferrer noopener'
          href='https://www.bajus.org/gold-price'
        >
          Bangladesh Jewellers Association website
        </a>
        <br />
        * There is a 5% VAT on all gold purchases in Bangladesh <br />* If purchased in
        jewelry form, there is additional making charges
      </p>

    </div>
  );
}

export default App;
