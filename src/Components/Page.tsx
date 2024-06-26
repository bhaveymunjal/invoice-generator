import { FC, ReactNode } from 'react'
import { Page as PdfPage } from '@react-pdf/renderer'
import compose from '../styles/compose'
import cn from 'classnames'

interface Props {
  className?: string
  pdfMode?: boolean
  children?: ReactNode
}

const Page: FC<Props> = ({ className, pdfMode, children }) => {
  return (
    <>
      {pdfMode ? (
        <PdfPage size="A4" style={compose('page ' + (className ? className : ''))}>
          {children}
        </PdfPage>
      ) : (
        <div className={cn('page', className)}>{children}</div>
      )}
    </>
  )
}

export default Page
