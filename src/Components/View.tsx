import React, { FC } from 'react'
import { View as PdfView } from '@react-pdf/renderer'
import compose from '../styles/compose'
import cn from 'classnames'

interface Props {
  className?: string
  pdfMode?: boolean
  children?: React.ReactNode
}

const View: FC<Props> = ({ className, pdfMode, children }) => {
  return (
    <>
      {pdfMode ? (
        <PdfView style={compose('view ' + (className ? className : ''))}>{children}</PdfView>
      ) : (
        <div className={cn('view', className)}>{children}</div>
      )}
    </>
  )
}

export default View
