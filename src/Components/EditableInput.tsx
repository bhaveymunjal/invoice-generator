import { FC } from 'react'
import { Text } from '@react-pdf/renderer'
import compose from '../styles/compose'
import cn from 'classnames'

interface Props {
  className?: string
  placeholder?: string
  value?: string | number
  onChange?: (value: string) => void
  pdfMode?: boolean
}

const EditableInput: FC<Props> = ({ className, placeholder, value, onChange, pdfMode }) => {
  return (
    <>
      {pdfMode ? (
        <Text style={compose('span ' + (className ? className : ''))}>{value}</Text>
      ) : (
        <input
          type="text"
          className={cn('input', className)}
          placeholder={placeholder || ''}
          value={value || ''}
          onChange={onChange ? (e) => onChange(e.target.value) : undefined}
        />
      )}
    </>
  )
}

export default EditableInput
